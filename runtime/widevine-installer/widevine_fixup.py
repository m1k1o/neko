#!/usr/bin/python3

"""
MIT License

Copyright (c) 2023 David Buchanan

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

Original script:
https://gist.github.com/DavidBuchanan314/c6b97add51b97e4c3ee95dc890f9e3c8
"""

import sys

verbose = False
args = sys.argv[1:]
if len(args) >= 1 and args[0] == "-v":
    verbose = True
    args = args[1:]

if len(args) != 2:
    print(f"Usage: {sys.argv[0]} [-v] input.so output.so")
    exit()

def log(s):
    if verbose:
        print(s)

"""
Widevine fixup tool for aarch64 systems

Old aarch64 widevine builds currently only support 4k page sizes.

This script fixes that, by pre-padding the LOAD segments so that they meet
the alignment constraints required by the loader, and then fixing up the
relevant header offsets to keep the file valid.

It also injects two functions that are not exported from typical libgccs, into
the empty space at the end of the .text segment. This avoids any LD_PRELOAD
workarounds. (The injected functions are __aarch64_ldadd4_acq_rel
and __aarch64_swp4_acq_rel)

IMPORTANT NOTE: On systems with >4k page size (e.g. Apple Silicon devices),
using the resulting binary *significantly* weakens the security of your web
browser, in two ways. Firstly, it disables the RELRO security mitigation, and
secondly it creates a RWX mapping.

This script also adds the necessary GLIBC_ABI_DT_RELR version tag so that
current glibc versions can load the library without requiring any patches.

Newer Widevine versions do have 64K aligned segments, and do not need the
padding process. They also do not have the same security implications, so its
use is recommended. However, we still adjust the segment offsets to open up
space for adding the missing functions, and to insert the GLIBC_ABI_DT_RELR
version.

This process is fragile, and may not work as-is on future revisions of widevine.
"""

import ctypes

class Elf64_Ehdr(ctypes.Structure):
    _fields_ = [
        ('e_ident', ctypes.c_ubyte * 16),
        ('e_type', ctypes.c_uint16),
        ('e_machine', ctypes.c_uint16),
        ('e_version', ctypes.c_uint32),
        ('e_entry', ctypes.c_uint64),
        ('e_phoff', ctypes.c_uint64),
        ('e_shoff', ctypes.c_uint64),
        ('e_flags', ctypes.c_uint32),
        ('e_ehsize', ctypes.c_uint16),
        ('e_phentsize', ctypes.c_uint16),
        ('e_phnum', ctypes.c_uint16),
        ('e_shentsize', ctypes.c_uint16),
        ('e_shnum', ctypes.c_uint16),
        ('e_shstrndx', ctypes.c_uint16),
    ]

class Elf64_Phdr(ctypes.Structure):
    _fields_ = [
        ('p_type', ctypes.c_uint32),
        ('p_flags', ctypes.c_uint32),
        ('p_offset', ctypes.c_uint64),
        ('p_vaddr', ctypes.c_uint64),
        ('p_paddr', ctypes.c_uint64),
        ('p_filesz', ctypes.c_uint64),
        ('p_memsz', ctypes.c_uint64),
        ('p_align', ctypes.c_uint64),
    ]

class P_FLAGS:
    """ Flag values for the p_flags field of program headers
    """
    PF_X=0x1
    PF_W=0x2
    PF_R=0x4
    PF_MASKOS=0x00FF0000
    PF_MASKPROC=0xFF000000

class PT:
    PT_NULL=0
    PT_LOAD=1
    PT_DYNAMIC=2
    PT_INTERP=3
    PT_NOTE=4
    PT_SHLIB=5
    PT_PHDR=6
    PT_TLS=7
    PT_LOOS=0x60000000
    PT_HIOS=0x6fffffff

    PT_GNU_EH_FRAME=0x6474e550
    PT_GNU_STACK=0x6474e551
    PT_GNU_RELRO=0x6474e552
    PT_GNU_PROPERTY=0x6474e553

class Elf64_Shdr(ctypes.Structure):
    _fields_ = [
        ('sh_name', ctypes.c_uint32),
        ('sh_type', ctypes.c_uint32),
        ('sh_flags', ctypes.c_uint64),
        ('sh_addr', ctypes.c_uint64),
        ('sh_offset', ctypes.c_uint64),
        ('sh_size', ctypes.c_uint64),
        ('sh_link', ctypes.c_uint32),
        ('sh_info', ctypes.c_uint32),
        ('sh_addralign', ctypes.c_uint64),
        ('sh_entsize', ctypes.c_uint64),
    ]

class Elf64_Sym(ctypes.Structure):
    _fields_ = [
        ('st_name', ctypes.c_uint32),
        ('st_info', ctypes.c_ubyte),
        ('st_other', ctypes.c_ubyte),
        ('st_shndx', ctypes.c_uint16),
        ('st_value', ctypes.c_uint64),
        ('st_size', ctypes.c_uint64),
    ]

class Elf64_Dyn(ctypes.Structure):
    _fields_ = [
        ('d_tag', ctypes.c_uint64),
        ('d_val', ctypes.c_uint64), # union with d_ptr
    ]

class D_TAG: # XXX: this is very incomplete
    DT_NULL=0
    DT_NEEDED=1
    DT_STRTAB=5
    DT_SONAME=14
    DT_VERNEED=0x6ffffffe

class Elf64_Rela(ctypes.Structure):
    _fields_ = [
        ('r_offset', ctypes.c_uint64),
        #('r_info', ctypes.c_uint64),
        ('r_type', ctypes.c_uint32),
        ('r_symbol', ctypes.c_uint32),
        ('r_addend', ctypes.c_int64),
    ]

class Elf64_Verneed(ctypes.Structure):
    _fields_ = [
        ('vn_version', ctypes.c_uint16),
        ('vn_cnt', ctypes.c_uint16),
        ('vn_file', ctypes.c_uint32),
        ('vn_aux', ctypes.c_uint32),
        ('vn_next', ctypes.c_uint32),
    ]

class Elf64_Vernaux(ctypes.Structure):
    _fields_ = [
        ('vna_hash', ctypes.c_uint32),
        ('vna_flags', ctypes.c_uint16),
        ('vna_other', ctypes.c_uint16),
        ('vna_name', ctypes.c_uint32),
        ('vna_next', ctypes.c_uint32),
    ]

import mmap
TARGET_PAGE_SIZE = mmap.PAGESIZE
WEAKEN_SECURITY = mmap.PAGESIZE > 0x1000
inject_addr = None

weakened_security = False

"""
0000000000000b24 <__aarch64_ldadd4_acq_rel>:
b24:   2a0003e2        mov     w2, w0
b28:   885ffc20        ldaxr   w0, [x1]
b2c:   0b020003        add     w3, w0, w2
b30:   8804fc23        stlxr   w4, w3, [x1]
b34:   35ffffa4        cbnz    w4, b28 <__aarch64_ldadd4_acq_rel+0x4>
b38:   d65f03c0        ret

0000000000000b3c <__aarch64_swp4_acq_rel>:
b3c:   2a0003e2        mov     w2, w0
b40:   885ffc20        ldaxr   w0, [x1]
b44:   8803fc22        stlxr   w3, w2, [x1]
b48:   35ffffc3        cbnz    w3, b40 <__aarch64_swp4_acq_rel+0x4>
b4c:   d65f03c0        ret
"""

injected_code = bytes.fromhex("e203002a20fc5f880300020b23fc0488a4ffff35c0035fd6e203002a20fc5f8822fc0388c3ffff35c0035fd6")

with open(args[0], "rb") as infile:
    elf = bytearray(infile.read())

print(f"Fixing up ChromeOS Widevine CDM module for Linux compatibility...")

elf_length = len(elf)
elf += bytearray(0x100000) # pre-expand the buffer by more than enough

ehdr = Elf64_Ehdr.from_buffer(elf)

phdrs = [
    Elf64_Phdr.from_buffer(memoryview(elf)[ehdr.e_phoff + i * ehdr.e_phentsize:])
    for i in range(ehdr.e_phnum)
]

adjustments = []

def adjust_offset(x):
    for index, delta in adjustments:
        if x >= index:
            x += delta
    return x

def align(a, b):
    return (a + b - 1) & ~(b - 1)

prev = None
remove_relro = False
for phdr in phdrs:
    phdr.p_offset = adjust_offset(phdr.p_offset)
    if phdr.p_type == PT.PT_DYNAMIC:
        phdr_dynamic = phdr
    if phdr.p_type == PT.PT_LOAD:
        if phdr.p_align < TARGET_PAGE_SIZE:
            phdr.p_align = TARGET_PAGE_SIZE
        delta_needed = (phdr.p_vaddr - phdr.p_offset) % phdr.p_align
        skip_perms_hack = False
        if phdr.p_vaddr != phdr.p_offset and not inject_addr:
            # Newer CDM versions use 64K alignment, so no longer require this hack.
            # However, we still need space to inject the code & modified headers, so
            # we might as well still do it.
            delta_needed = phdr.p_vaddr - phdr.p_offset
            skip_perms_hack = True
        if delta_needed:
            log(f"  Inserting {hex(delta_needed)} bytes at offset {hex(phdr.p_offset)}")
            if not inject_addr:
                pad_bytes = injected_code + bytes(delta_needed - len(injected_code))
                inject_addr = phdr.p_offset
                log(f"  Also injecting code at offset {hex(phdr.p_offset)}")
            else:
                pad_bytes = bytes(delta_needed)
            elf[phdr.p_offset:] = pad_bytes + elf[phdr.p_offset:-delta_needed]
            adjustments.append((phdr.p_offset, delta_needed))
            elf_length += delta_needed
            phdr.p_offset += delta_needed

            # Load the injected bytes up to the align size as part of the previous phdr
            align_off = align(prev.p_vaddr + prev.p_filesz, prev.p_align) - prev.p_vaddr - prev.p_filesz
            # This could fail if we get unlucky, let's hope not
            assert align_off >= len(injected_code)
            prev.p_filesz += min(delta_needed, align_off)
            prev.p_memsz += min(delta_needed, align_off)

            if WEAKEN_SECURITY and not skip_perms_hack:
                phdr.p_flags |= P_FLAGS.PF_X # XXX: this is a hack!!! (at the very least, we should apply it only to the mappings that need it)
                remove_relro = True
                weakened_security = True
    prev = phdr

    if WEAKEN_SECURITY and remove_relro and phdr.p_type == PT.PT_GNU_RELRO:
        print("  Neutering relro") # XXX: relro is a security mechanism
        phdr.p_type = PT.PT_NOTE
        weakened_security = True

if inject_addr is None:
    inject_addr = (elf_length + 3) & ~3
    elf[inject_addr: inject_addr + len(injected_code)] = injected_code
    elf_length += 0x10000


free_addr = inject_addr + len(injected_code)
# the section headers have moved
ehdr.e_shoff = adjust_offset(ehdr.e_shoff)

shdrs = [
    Elf64_Shdr.from_buffer(memoryview(elf)[ehdr.e_shoff + i * ehdr.e_shentsize:])
    for i in range(ehdr.e_shnum)
]

for shdr in shdrs:
    shdr.sh_offset = adjust_offset(shdr.sh_offset)

strtab = shdrs[ehdr.e_shstrndx]

def resolve_string(elf, strtab, stridx, count=False):
    if count:
        str_start = strtab.sh_offset
        for _ in range(stridx):
            str_start = elf.index(b"\0", str_start) + 1
    else:
        str_start = strtab.sh_offset + stridx
    str_end = elf.index(b"\0", str_start)
    return bytes(elf[str_start:str_end])

shdr_by_name = {
    resolve_string(elf, strtab, shdr.sh_name): shdr
    for shdr in shdrs
}

# XXX: unfortunately this does not do anything useful!
# It doesn't hurt either, so I'm leaving it here just in case.
dynsym = shdr_by_name[b".dynsym"]
dynstr = shdr_by_name[b".dynstr"]
for i in range(0, dynsym.sh_size, dynsym.sh_entsize):
    sym = Elf64_Sym.from_buffer(memoryview(elf)[dynsym.sh_offset + i:])
    name = resolve_string(elf, dynstr, sym.st_name)
    if name in [b"__aarch64_ldadd4_acq_rel", b"__aarch64_swp4_acq_rel"]:
        log(f"  Weak binding {name}")
        sym.st_info = (sym.st_info & 0x0f) | (2 << 4) # STB_WEAK

"""
dynamic = shdr_by_name[b".dynamic"]
for i in range(0, dynamic.sh_size, dynamic.sh_entsize):
    dyn = Elf64_Dyn.from_buffer(memoryview(elf)[dynamic.sh_offset + i:])
    if dyn.d_tag == D_TAG.DT_SONAME:
        print("hijacking SONAME tag to point to NEEDED libgcc_hide.so")
        dyn.d_tag = D_TAG.DT_NEEDED
        dyn.d_val = inject_addr - dynstr.sh_offset
        dynstr.sh_size = (inject_addr - dynstr.sh_offset) + len(PATH_TO_INJECT) + 1
"""

rela_plt = shdr_by_name[b".rela.plt"]
for i in range(0, rela_plt.sh_size, rela_plt.sh_entsize):
    rela = Elf64_Rela.from_buffer(memoryview(elf)[rela_plt.sh_offset + i:])
    sym = resolve_string(elf, dynstr, rela.r_symbol, count=True)
    if sym in [b"__aarch64_ldadd4_acq_rel", b"__aarch64_swp4_acq_rel"]:
        log(f"  Modifying {sym} plt reloc to point into injected code")
        rela.r_type = 1027 # R_AARCH64_RELATIVE
        rela.r_addend = inject_addr
        if sym == b"__aarch64_swp4_acq_rel":
            rela.r_addend += 6*4

# Move the dynstr section to the hole and add the missing GLIBC_ABI_DT_RELR
log("  Moving .dynstr to free space and adding GLIBC_ABI_DT_RELR...")
free_addr = (free_addr + 3) & ~3
dynstr = shdr_by_name[b".dynstr"]
dynstr_data = elf[dynstr.sh_offset:dynstr.sh_offset + dynstr.sh_size]
abi_dt_relr_off = len(dynstr_data)
dynstr_data += b"GLIBC_ABI_DT_RELR\0"
dynstr.sh_offset = free_addr
dynstr.sh_addr = free_addr
dynstr.sh_size = len(dynstr_data)
elf[free_addr:free_addr + dynstr.sh_size] = dynstr_data
free_addr += dynstr.sh_size

log("  Moving .gnu.version_r to free space and adding GLIBC_ABI_DT_RELR...")
ver_r = shdr_by_name[b".gnu.version_r"]
ver_r_data = elf[ver_r.sh_offset:ver_r.sh_offset + ver_r.sh_size]
# We need one more vernaux entry
ver_r_data += bytes(16)

p = 0
offset = 0
while True:
    need = Elf64_Verneed.from_buffer(memoryview(ver_r_data)[p: p + 16])
    filename = resolve_string(elf, dynstr, need.vn_file)
    need.vn_aux += offset
    if filename == b'libc.so.6':
        q = p + need.vn_aux
        for i in range(need.vn_cnt):
            aux = Elf64_Vernaux.from_buffer(memoryview(ver_r_data)[q: q + 16])
            ver = resolve_string(elf, dynstr, aux.vna_name)
            q += aux.vna_next
        need.vn_cnt += 1
        aux.vna_next = 16
        q += 16
        # Make space here
        ver_r_data[q + 16:] = ver_r_data[q:-16]
        aux = Elf64_Vernaux.from_buffer(memoryview(ver_r_data)[q: q + 16])
        aux.vna_hash = 0xfd0e42
        aux.vna_name = abi_dt_relr_off
        aux.vna_other = 3

        # Shift the rest of the aux offsets
        offset = 16

    if need.vn_next == 0:
        break
    p += need.vn_next

free_addr = (free_addr + 3) & ~3
ver_r.sh_offset = free_addr
ver_r.sh_addr = free_addr
ver_r.sh_size = len(ver_r_data)
elf[free_addr:free_addr + ver_r.sh_size] = ver_r_data
free_addr += ver_r.sh_size

# Now fix the DYNAMIC section
log("  Fixing up DYNAMIC section...")
for p in range(phdr_dynamic.p_offset, phdr_dynamic.p_offset + phdr_dynamic.p_filesz, 16):
    dyn = Elf64_Dyn.from_buffer(memoryview(elf)[p: p + 16])
    if dyn.d_tag == D_TAG.DT_VERNEED:
        dyn.d_val = ver_r.sh_offset
    if dyn.d_tag == D_TAG.DT_STRTAB:
        dyn.d_val = dynstr.sh_offset

if not weakened_security:
    print()
    print("Good news! This CDM version supports your page size, so we didn't have")
    print("to weaken memory permissions. Rejoice!")
else:
    print()
    print("It looks like you're running Asahi, or some other device with >4k page size.")
    print("This CDM only supports smaller page sizes, so we had to weaken memory")
    print("permissions to make it work.")

with open(args[1], "wb") as outfile:
    outfile.write(memoryview(elf)[:elf_length])

