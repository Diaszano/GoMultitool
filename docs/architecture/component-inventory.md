# Upstream component inventory

Inventory at baseline commit `d7dc392fe33850f5975aa466538fd889629c430c`.

| Path | Contents |

|---|---|

| `create_image.sh` | Builds a board-specific recovery image using host image and filesystem tools. |
| `sources/multistrap/` | Root filesystem bootstrap, init, menu, hooks, and overlays. |
| `sources/rk322x*` | RK322x board configuration, Linux configuration, device tree, bootloader, and TrustZone firmware. |
| `sources/rk3288*`, `sources/rk3318*`, `sources/rk3528*` | Other board configurations and their boot components. |
| `tools/` | Bundled Rockchip image and firmware utilities (`loaderimage`, `mkimage`, `resource_tool`, `rkcrc`). |
| `LICENSE` | Repository GPL v2 license text; individual assets may carry additional terms. |

## Tracked source files
The following list is generated from the baseline commit so it can be compared directly with Git:

```text
create_image.sh
sources/multistrap/hooks/completion01.sh
sources/multistrap/init
sources/multistrap/multistrap.conf
sources/multistrap/multitool.sh
sources/multistrap/overlays/etc/chrony/chrony.conf
sources/multistrap/overlays/etc/dropbear/banner
sources/multistrap/overlays/etc/dropbear/dropbear_dss_host_key
sources/multistrap/overlays/etc/dropbear/dropbear_dss_host_key.pub
sources/multistrap/overlays/etc/dropbear/dropbear_ecdsa_host_key
sources/multistrap/overlays/etc/dropbear/dropbear_ecdsa_host_key.pub
sources/multistrap/overlays/etc/dropbear/dropbear_rsa_host_key
sources/multistrap/overlays/etc/dropbear/dropbear_rsa_host_key.pub
sources/multistrap/overlays/etc/fstab
sources/multistrap/overlays/etc/hostname
sources/rk322x.conf
sources/rk322x/assets/linux.config
sources/rk322x/assets/stepnand-patches/readme.txt
sources/rk322x/assets/stepnand-patches/u-boot-0002-rockchip-board-use-efuse-cpuid-to-set-ethaddr.patch
sources/rk322x/assets/stepnand-patches/u-boot-0003-rockchip-rk3328-get-serial-and-ethaddr-from-efuse.patch
sources/rk322x/assets/stepnand-patches/u-boot-0004-rk3328-evb-fix-sdmmc-vmmc-supply.patch
sources/rk322x/assets/stepnand-patches/u-boot-0005-rockchip-rk3288-get-serial-and-ethaddr-from-efuse.patch
sources/rk322x/assets/stepnand-patches/u-boot-0006-rk3399-evb-use-common-get-serial-and-ethaddr-from-ef.patch
sources/rk322x/assets/stepnand-patches/u-boot-0007-include-update-log2-header-from-the-Linux-kernel.patch
sources/rk322x/assets/stepnand-patches/u-boot-0008-rockchip-disable-android-boot-and-config.patch
sources/rk322x/assets/stepnand-patches/u-boot-0009-configs-sync-rk3288-rk3328-and-rk3399-defconfig.patch
sources/rk322x/assets/stepnand-patches/u-boot-0010-rk3288-tinker-enable-emmc.patch
sources/rk322x/assets/stepnand-patches/u-boot-0011-Revert-rockchip-dts-rk3328-evb-Enable-gmac2io-for-rk.patch
sources/rk322x/assets/stepnand-patches/u-boot-0012-Makefile-compat.patch
sources/rk322x/assets/stepnand-patches/u-boot-1000-rk322x-defconfig-dts.patch
sources/rk322x/assets/stepnand-patches/u-boot-2000-rk322x-add-dtsi-generic-tvbox.patch
sources/rk322x/assets/stepnand-patches/u-boot-2001-rk322x-box-add-defconfig.patch
sources/rk322x/assets/stepnand-patches/u-boot-2002-rk322x-box-add-dts-makefile.patch
sources/rk322x/assets/stepnand-patches/u-boot-2003-rk322x-box-add-dts.patch
sources/rk322x/assets/stepnand-patches/u-boot-2004-rk322x-box-add-nand-env-boot.patch
sources/rk322x/boot_install
sources/rk322x/extlinux.conf
sources/rk322x/idbloader.img
sources/rk322x/idbloader_ddr_1.11.img
sources/rk322x/kernel.img
sources/rk322x/legacy-rk-uboot.bin
sources/rk322x/rk3228_tee_ta-51.1.0-333-gc9d95d1.bin
sources/rk322x/rk322x-box.dtb
sources/rk322x/rk322x_optee_os_3.7.0.bin
sources/rk322x/rk322x_optee_os_4.5.0.bin
sources/rk322x/rk322x_tee_1.0.1-65.bin
sources/rk322x/rk322x_tee_1.0.1-72-gf230aa2.bin
sources/rk322x/rk322x_tee_ta_1.1.0-120.bin
sources/rk322x/rk322x_tee_ta_1.1.0-297-ga4fd2d1.bin
sources/rk322x/u-boot-dtb.bin
sources/rk3288.conf
sources/rk3288/assets/linux.config
sources/rk3288/boot_install
sources/rk3288/extlinux.conf
sources/rk3288/idbloader.img
sources/rk3288/initrd.img
sources/rk3288/kernel-5.6.img
sources/rk3288/kernel.img
sources/rk3288/linux-kernel.config
sources/rk3288/parameter-linux.img
sources/rk3288/rk3288-box-5.6.dtb
sources/rk3288/rk3288-box-5.6.dts
sources/rk3288/rk3288-box.dtb
sources/rk3288/rk3288_tee_ta_v1.42.bin
sources/rk3288/u-boot-dtb.bin
sources/rk3318.conf
sources/rk3318/assets/RK3328TRUST.ini
sources/rk3318/assets/bl31.elf
sources/rk3318/assets/kernel.config
sources/rk3318/assets/rk-uboot.config
sources/rk3318/assets/rk-uboot.txt
sources/rk3318/assets/rk3318-box-uboot.dts
sources/rk3318/assets/rk3318-box-uboot.dtsi
sources/rk3318/assets/rk3318-box.dts
sources/rk3318/assets/u-boot.config
sources/rk3318/boot_install
sources/rk3318/extlinux.conf
sources/rk3318/idbloader.img
sources/rk3318/kernel.img
sources/rk3318/legacy-rk-uboot.bin
sources/rk3318/rk3318-box.dtb
sources/rk3318/trustos.bin
sources/rk3318/u-boot.bin
sources/rk3318/u-boot.config
sources/rk3528.conf
sources/rk3528/assets/kernel-sources.txt
sources/rk3528/assets/linux.config
sources/rk3528/assets/rk3528-box.dts
sources/rk3528/assets/rk3528_bl31_v1.14.elf
sources/rk3528/assets/rk3528_bl32_v1.02.bin
sources/rk3528/assets/rk3528_ddr_1056MHz_v1.05.bin
sources/rk3528/assets/rk3528_spl_v1.04.bin
sources/rk3528/assets/u-boot.config
sources/rk3528/boot_install
sources/rk3528/extlinux.conf
sources/rk3528/idbloader.img
sources/rk3528/kernel.img
sources/rk3528/rk3528-box.dtb
sources/rk3528/u-boot.itb
tools/loaderimage
tools/mkimage
tools/resource_tool
tools/rkcrc
```
