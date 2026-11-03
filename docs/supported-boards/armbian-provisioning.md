# Armbian offline provisioning

The project profile accepts Armbian images with `ID=armbian` and version `26.08` or newer. The first-boot artifact uses the documented `/root/.not_logged_in_yet` format and preset keys for root password, user name/password, real name, locale, timezone, Ethernet, and network-default behavior. See the [Armbian first-boot configuration reference](https://docs.armbian.com/user-guide/autoconfig/) for the current directives.

Generated first-boot files use mode `0600` because Armbian stores password values in plaintext. The image modification step must keep temporary image copies private and remove them on failure. An optional SSH public key is placed in the new user's `authorized_keys`; this does not enable an SSH daemon if the image does not already provide one.

The schema has been checked against Armbian's published configuration reference. The image editor currently accepts regular image files with an MBR table containing an ext4 root partition (or a raw ext4 filesystem); other layouts are rejected. Mounting uses `losetup` and `mount` and requires Linux root privileges. No physical Armbian image was available here, so a target image's exact first-boot behavior and RK3229 bootability remain unverified. Unsupported OS IDs, versions below the profile floor, and missing preset directives are rejected.
