# multitool
SD/eMMC card helper Multitool for TV boxes and alike

# Building

This assumes a Debian-derived system, adapt to your system as appropriate. First, install required packages:
```sh
sudo apt install multistrap squashfs-tools parted dosfstools ntfs-3g qemu-user-static debian-keyring debian-archive-keyring
```

Fetch the source code:
```sh
git clone https://github.com/paolosabatino/multitool
```

Then build an image for the appropriate board (root is required, as the script must have enough permissions to set up and manipulate loop devices for the target image):
```sh
sudo ./create_image.sh $board
```
See `sources/*.conf` for supported board configurations; the resulting image can be found at `dist-$board/multitool.img`.

## Preserving the original recovery system

The original Bash recovery menu, board assets, and image builder remain available while the Go application is developed. See [the upstream baseline](docs/architecture/upstream-baseline.md) and [component inventory](docs/architecture/component-inventory.md) before changing boot components.
