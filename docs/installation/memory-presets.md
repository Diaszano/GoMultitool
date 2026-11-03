# Optional memory presets

Swap is opt-in. The installer must confirm the requested size and available image space before adding the generated systemd swap unit. It refuses to add the preset when a known ZRAM manager is already installed, avoiding competing memory services. The configured priority is preserved so the kernel can order swap devices.

The preset does not create or enable a swap file on the running host. The image build workflow is responsible for creating the file only after explicit user authorization. Swap writes increase eMMC wear; prefer existing ZRAM on eMMC systems when appropriate.
