# RK3229 reference board

The initial target is the MXQ Pro 4K TV box with Rockchip RK3229, nominally 1 GiB RAM and 8 GB eMMC, booted from microSD. These are the design target recorded in the modernization specification, not measurements from a board attached during this work.

## Repository boot inputs

`sources/rk322x.conf` selects `rk322x-box.dtb`, `idbloader_ddr_1.11.img`, `u-boot-dtb.bin`, `kernel.img`, and `rk322x_optee_os_4.5.0.bin`. Other legacy variants are retained in `sources/rk322x/`; do not substitute them without board validation.

SHA-256 checksums (computed from this checkout):

| File | SHA-256 |
| --- | --- |
| `sources/rk322x/idbloader_ddr_1.11.img` | `0fc2ffec025e1a69ea46f0e00d48ca99d26b61e372e2dbf533937b8d91a1444a` |
| `sources/rk322x/u-boot-dtb.bin` | `f09cbe018107a32ff27b560a555d555d271e2c299ddada99030890fc2a6d9630` |
| `sources/rk322x/kernel.img` | `d93588d737ed3a891725ad89cfb563346bbc961cd8a776a9f371bc7026d6831d` |
| `sources/rk322x/rk322x-box.dtb` | `1726b4706d2ab4a401588578fd976bbcda19e2baef824f9dae1e06f6f316f9b5` |
| `sources/rk322x/rk322x_optee_os_4.5.0.bin` | `0fc2ffec025e1a69ea46f0e00d48ca99d26b61e372e2dbf533937b8d91a1444a` |

These checksums provide reproducible source identity only. The complete recovery image is not present in this repository, so its checksum is unavailable.

## Validation status and limitations

- Source configuration and tracked component files have been inspected; a physical RK3229 board was not available for this baseline review.
- Boot from microSD, HDMI, keyboard, RAM size, eMMC identity, and recovery behavior remain unverified on hardware.
- RK322x covers multiple board variants. The generic device tree and configuration do not establish compatibility with every RK322x device.
- Preserve the original recovery card and firmware. Test only with expendable media and require explicit human authorization before any physical write.

No image-writing operation should infer the destination from a fixed `/dev/mmcblk*` name.
