# Recovery and hardware status

The existing Bash recovery menu and board boot files remain the device recovery interface. The Go executable is currently a version-reporting command and is not a replacement recovery menu.

## Board status

| Board family | Repository status | New image pipeline status |
| --- | --- | --- |
| RK322x | Legacy configuration and boot assets are present. | Automated image structure checks are defined; physical boot compatibility is unverified. |
| RK3288 | Legacy configuration and assets are present. | Not built by the new image pipeline. |
| RK3318 | Legacy configuration and assets are present. | Not built by the new image pipeline. |
| RK3528 | Legacy configuration and assets are present. | Not built by the new image pipeline. |

A board configuration in the repository does not by itself establish tested hardware support. See the [component inventory](../architecture/component-inventory.md) for the files and the [firmware release policy](../releases/release-policy.md) for promotion requirements.

## Device and data risks

Recovery and boot tools can write to internal flash storage and make a device unbootable or erase user data. Confirm the board model and recovery method before using an image, keep a verified backup, and do not use a stable-support claim without documented hardware testing. The image build target creates and validates a regular file only; it does not flash a device.

No generic flash or unbrick procedure is provided here because boot modes and recovery methods vary by board and device. Use hardware-specific instructions and verify the target before any write operation.
