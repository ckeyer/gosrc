// Do not edit. Bootstrap copy of /usr/local/go/src/cmd/link/internal/ld/macho_combine_dwarf.go

//line /usr/local/go/src/cmd/link/internal/ld/macho_combine_dwarf.go:1
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"bytes"
	"debug/macho"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
	"unsafe"
)

var fakedwarf, realdwarf, linkseg *macho.Segment
var dwarfstart, linkstart int64
var linkoffset uint32
var machHeader *macho.FileHeader
var mappedHeader []byte

const (
	LC_ID_DYLIB             = 0xd
	LC_LOAD_DYLINKER        = 0xe
	LC_PREBOUND_DYLIB       = 0x10
	LC_LOAD_WEAK_DYLIB      = 0x18
	LC_UUID                 = 0x1b
	LC_RPATH                = 0x8000001c
	LC_CODE_SIGNATURE       = 0x1d
	LC_SEGMENT_SPLIT_INFO   = 0x1e
	LC_REEXPORT_DYLIB       = 0x8000001f
	LC_ENCRYPTION_INFO      = 0x21
	LC_DYLD_INFO            = 0x22
	LC_DYLD_INFO_ONLY       = 0x80000022
	LC_VERSION_MIN_MACOSX   = 0x24
	LC_VERSION_MIN_IPHONEOS = 0x25
	LC_FUNCTION_STARTS      = 0x26
	LC_MAIN                 = 0x80000028
	LC_DATA_IN_CODE         = 0x29
	LC_SOURCE_VERSION       = 0x2A
	LC_DYLIB_CODE_SIGN_DRS  = 0x2B
	LC_ENCRYPTION_INFO_64   = 0x2C

	dwarfMinAlign = 6  // 64 = 1 << 6
	pageAlign     = 12 // 4096 = 1 << 12
)

type loadCmd struct {
	Cmd macho.LoadCmd
	Len uint32
}

type dyldInfoCmd struct {
	Cmd                      macho.LoadCmd
	Len                      uint32
	RebaseOff, RebaseLen     uint32
	BindOff, BindLen         uint32
	WeakBindOff, WeakBindLen uint32
	LazyBindOff, LazyBindLen uint32
	ExportOff, ExportLen     uint32
}

type linkEditDataCmd struct {
	Cmd              macho.LoadCmd
	Len              uint32
	DataOff, DataLen uint32
}

type encryptionInfoCmd struct {
	Cmd                macho.LoadCmd
	Len                uint32
	CryptOff, CryptLen uint32
	CryptId            uint32
}

type loadCmdReader struct {
	offset, next int64
	f            *os.File
	order        binary.ByteOrder
}

func (r *loadCmdReader) Next() (cmd loadCmd, err error) {
	r.offset = r.next
	if _, err = r.f.Seek(r.offset, 0); err != nil {
		return
	}
	if err = binary.Read(r.f, r.order, &cmd); err != nil {
		return
	}
	r.next = r.offset + int64(cmd.Len)
	return
}

func (r loadCmdReader) ReadAt(offset int64, data interface{}) error {
	if _, err := r.f.Seek(r.offset+offset, 0); err != nil {
		return err
	}
	return binary.Read(r.f, r.order, data)
}

func (r loadCmdReader) WriteAt(offset int64, data interface{}) error {
	if _, err := r.f.Seek(r.offset+offset, 0); err != nil {
		return err
	}
	return binary.Write(r.f, r.order, data)
}

// machoCombineDwarf merges dwarf info generated by dsymutil into a macho executable.
// With internal linking, DWARF is embedded into the executable, this lets us do the
// same for external linking.
// inexe is the path to the executable with no DWARF. It must have enough room in the macho
// header to add the DWARF sections. (Use ld's -headerpad option)
// dsym is the path to the macho file containing DWARF from dsymutil.
// outexe is the path where the combined executable should be saved.
func machoCombineDwarf(inexe, dsym, outexe string) error {
	exef, err := os.Open(inexe)
	if err != nil {
		return err
	}
	dwarff, err := os.Open(dsym)
	if err != nil {
		return err
	}
	outf, err := os.Create(outexe)
	if err != nil {
		return err
	}
	outf.Chmod(0755)

	exem, err := macho.NewFile(exef)
	if err != nil {
		return err
	}
	dwarfm, err := macho.NewFile(dwarff)
	if err != nil {
		return err
	}

	// The string table needs to be the last thing in the file
	// for code signing to work. So we'll need to move the
	// linkedit section, but all the others can be copied directly.
	linkseg = exem.Segment("__LINKEDIT")
	if linkseg == nil {
		return fmt.Errorf("missing __LINKEDIT segment")
	}

	if _, err = exef.Seek(0, 0); err != nil {
		return err
	}
	if _, err := io.CopyN(outf, exef, int64(linkseg.Offset)); err != nil {
		return err
	}

	realdwarf = dwarfm.Segment("__DWARF")
	if realdwarf == nil {
		return fmt.Errorf("missing __DWARF segment")
	}

	// Now copy the dwarf data into the output.
	maxalign := uint32(dwarfMinAlign) //
	for _, sect := range dwarfm.Sections {
		if sect.Align > maxalign {
			maxalign = sect.Align
		}
	}
	dwarfstart = machoCalcStart(realdwarf.Offset, linkseg.Offset, maxalign)
	if _, err = outf.Seek(dwarfstart, 0); err != nil {
		return err
	}

	if _, err = dwarff.Seek(int64(realdwarf.Offset), 0); err != nil {
		return err
	}
	if _, err := io.CopyN(outf, dwarff, int64(realdwarf.Filesz)); err != nil {
		return err
	}

	// And finally the linkedit section.
	if _, err = exef.Seek(int64(linkseg.Offset), 0); err != nil {
		return err
	}
	linkstart = machoCalcStart(linkseg.Offset, uint64(dwarfstart)+realdwarf.Filesz, pageAlign)
	linkoffset = uint32(linkstart - int64(linkseg.Offset))
	if _, err = outf.Seek(linkstart, 0); err != nil {
		return err
	}
	if _, err := io.Copy(outf, exef); err != nil {
		return err
	}

	// Now we need to update the headers.
	cmdOffset := unsafe.Sizeof(exem.FileHeader)
	is64bit := exem.Magic == macho.Magic64
	if is64bit {
		// mach_header_64 has one extra uint32.
		cmdOffset += unsafe.Sizeof(exem.Magic)
	}

	textsect := exem.Section("__text")
	if linkseg == nil {
		return fmt.Errorf("missing __text section")
	}

	dwarfCmdOffset := int64(cmdOffset) + int64(exem.FileHeader.Cmdsz)
	availablePadding := int64(textsect.Offset) - dwarfCmdOffset
	if availablePadding < int64(realdwarf.Len) {
		return fmt.Errorf("No room to add dwarf info. Need at least %d padding bytes, found %d", realdwarf.Len, availablePadding)
	}
	// First, copy the dwarf load command into the header
	if _, err = outf.Seek(dwarfCmdOffset, 0); err != nil {
		return err
	}
	if _, err := io.CopyN(outf, bytes.NewReader(realdwarf.Raw()), int64(realdwarf.Len)); err != nil {
		return err
	}

	if _, err = outf.Seek(int64(unsafe.Offsetof(exem.FileHeader.Ncmd)), 0); err != nil {
		return err
	}
	if err = binary.Write(outf, exem.ByteOrder, exem.Ncmd+1); err != nil {
		return err
	}
	if err = binary.Write(outf, exem.ByteOrder, exem.Cmdsz+realdwarf.Len); err != nil {
		return err
	}

	reader := loadCmdReader{next: int64(cmdOffset), f: outf, order: exem.ByteOrder}
	for i := uint32(0); i < exem.Ncmd; i++ {
		cmd, err := reader.Next()
		if err != nil {
			return err
		}
		switch cmd.Cmd {
		case macho.LoadCmdSegment64:
			err = machoUpdateSegment(reader, &macho.Segment64{}, &macho.Section64{})
		case macho.LoadCmdSegment:
			err = machoUpdateSegment(reader, &macho.Segment32{}, &macho.Section32{})
		case LC_DYLD_INFO, LC_DYLD_INFO_ONLY:
			err = machoUpdateLoadCommand(reader, &dyldInfoCmd{}, "RebaseOff", "BindOff", "WeakBindOff", "LazyBindOff", "ExportOff")
		case macho.LoadCmdSymtab:
			err = machoUpdateLoadCommand(reader, &macho.SymtabCmd{}, "Symoff", "Stroff")
		case macho.LoadCmdDysymtab:
			err = machoUpdateLoadCommand(reader, &macho.DysymtabCmd{}, "Tocoffset", "Modtaboff", "Extrefsymoff", "Indirectsymoff", "Extreloff", "Locreloff")
		case LC_CODE_SIGNATURE, LC_SEGMENT_SPLIT_INFO, LC_FUNCTION_STARTS, LC_DATA_IN_CODE, LC_DYLIB_CODE_SIGN_DRS:
			err = machoUpdateLoadCommand(reader, &linkEditDataCmd{}, "DataOff")
		case LC_ENCRYPTION_INFO, LC_ENCRYPTION_INFO_64:
			err = machoUpdateLoadCommand(reader, &encryptionInfoCmd{}, "CryptOff")
		case macho.LoadCmdDylib, macho.LoadCmdThread, macho.LoadCmdUnixThread, LC_PREBOUND_DYLIB, LC_UUID, LC_VERSION_MIN_MACOSX, LC_VERSION_MIN_IPHONEOS, LC_SOURCE_VERSION, LC_MAIN, LC_LOAD_DYLINKER, LC_LOAD_WEAK_DYLIB, LC_REEXPORT_DYLIB, LC_RPATH, LC_ID_DYLIB:
			// Nothing to update
		default:
			err = fmt.Errorf("Unknown load command 0x%x (%s)\n", int(cmd.Cmd), cmd.Cmd)
		}
		if err != nil {
			return err
		}
	}
	return machoUpdateDwarfHeader(&reader)
}

// machoUpdateSegment updates the load command for a moved segment.
// Only the linkedit segment should move, and it should have 0 sections.
// seg should be a macho.Segment32 or macho.Segment64 as appropriate.
// sect should be a macho.Section32 or macho.Section64 as appropriate.
func machoUpdateSegment(r loadCmdReader, seg, sect interface{}) error {
	if err := r.ReadAt(0, seg); err != nil {
		return err
	}
	segValue := reflect.ValueOf(seg)
	offset := reflect.Indirect(segValue).FieldByName("Offset")

	// Only the linkedit segment moved, any thing before that is fine.
	if offset.Uint() < linkseg.Offset {
		return nil
	}
	offset.SetUint(offset.Uint() + uint64(linkoffset))
	if err := r.WriteAt(0, seg); err != nil {
		return err
	}
	// There shouldn't be any sections, but just to make sure...
	return machoUpdateSections(r, segValue, reflect.ValueOf(sect), uint64(linkoffset))
}

func machoUpdateSections(r loadCmdReader, seg, sect reflect.Value, delta uint64) error {
	iseg := reflect.Indirect(seg)
	nsect := iseg.FieldByName("Nsect").Uint()
	if nsect == 0 {
		return nil
	}
	sectOffset := int64(iseg.Type().Size())

	isect := reflect.Indirect(sect)
	offsetField := isect.FieldByName("Offset")
	reloffField := isect.FieldByName("Reloff")
	sectSize := int64(isect.Type().Size())
	for i := uint64(0); i < nsect; i++ {
		if err := r.ReadAt(sectOffset, sect.Interface()); err != nil {
			return err
		}
		if offsetField.Uint() != 0 {
			offsetField.SetUint(offsetField.Uint() + delta)
		}
		if reloffField.Uint() != 0 {
			reloffField.SetUint(reloffField.Uint() + delta)
		}
		if err := r.WriteAt(sectOffset, sect.Interface()); err != nil {
			return err
		}
		sectOffset += sectSize
	}
	return nil
}

// machoUpdateDwarfHeader updates the DWARF segment load command.
func machoUpdateDwarfHeader(r *loadCmdReader) error {
	var seg, sect interface{}
	cmd, err := r.Next()
	if err != nil {
		return err
	}
	if cmd.Cmd == macho.LoadCmdSegment64 {
		seg = new(macho.Segment64)
		sect = new(macho.Section64)
	} else {
		seg = new(macho.Segment32)
		sect = new(macho.Section32)
	}
	if err := r.ReadAt(0, seg); err != nil {
		return err
	}
	segValue := reflect.ValueOf(seg)
	offset := reflect.Indirect(segValue).FieldByName("Offset")

	delta := uint64(dwarfstart) - realdwarf.Offset
	offset.SetUint(offset.Uint() + delta)
	if err := r.WriteAt(0, seg); err != nil {
		return err
	}
	return machoUpdateSections(*r, segValue, reflect.ValueOf(sect), delta)
}

func machoUpdateLoadCommand(r loadCmdReader, cmd interface{}, fields ...string) error {
	if err := r.ReadAt(0, cmd); err != nil {
		return err
	}
	value := reflect.Indirect(reflect.ValueOf(cmd))

	for _, name := range fields {
		field := value.FieldByName(name)
		fieldval := field.Uint()
		if fieldval >= linkseg.Offset {
			field.SetUint(fieldval + uint64(linkoffset))
		}
	}
	if err := r.WriteAt(0, cmd); err != nil {
		return err
	}
	return nil
}

func machoCalcStart(origAddr, newAddr uint64, alignExp uint32) int64 {
	align := uint64(1 << alignExp)
	if (origAddr % align) == (newAddr % align) {
		return int64(newAddr)
	}
	padding := (align - (newAddr % align))
	padding += origAddr % align
	return int64(padding + newAddr)
}
