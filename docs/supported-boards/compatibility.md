# Board compatibility profiles

Profiles match exact device-tree compatible strings. A missing match remains unknown; multiple matches are rejected as ambiguous. Storage kinds absent from the profile are unsupported. Profile recognition does not prove that a device is physically compatible or authorize a write.

The RK3229 reference profile records SHA-256 values for the repository's selected `u-boot-dtb.bin` and `kernel.img`. These hashes identify local files; no upstream release version or physical boot result was established. Bootloader modification remains disabled for every profile until hardware compatibility is verified.

The bootloader inspector only computes SHA-256 and reports whether it matches a supplied known hash. It has no write operation. Preserve the existing bootloader and kernel during all tests.
