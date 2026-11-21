# Recovery image build

The recovery image builder currently supports the verified RK322x target and requires Linux, root privileges, loop devices, and the tools installed by the image workflow. Build it with:

```sh
make image
```

The builder cross-compiles the Go command for ARMv7, places it in the recovery root filesystem and data partition, then runs structural checks against the partition table, filesystems, boot files, and embedded ARM executables. It only attaches the generated image file to loop devices; it does not accept or write to a physical block device.

Successful builds produce an XZ-compressed image, SHA-256 checksum, and JSON manifest under `image-dist/`. The manifest records the source commit and SHA-256 digests for the kernel and bootloader inputs. The image workflow runs on pull requests, pushes to `main`, and manual dispatch. It uploads artifacts only after validation succeeds; it does not publish firmware releases.

The builder uses the host's `debian-archive-keyring` package to authenticate Debian's `trixie` repository at `https://deb.debian.org/debian/`. Hardware boot and recovery behavior still require validation on supported RK322x boards before an image can be considered stable.

The HDMI menu targets HD (1280×720) and scales its console font at Full HD (1920×1080) and UHD 4K (3840×2160). The kernel and display negotiate the video mode during boot. A 4K TV can display the menu at a lower mode if that is what the board and display negotiate. Below HD or when framebuffer geometry is unavailable, boot uses the existing console font.
