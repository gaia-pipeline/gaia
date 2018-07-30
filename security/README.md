# Security

## Certificates

Gaia, when first started will create a signed certificate in a location
defined by the user under `gaia.Cfg.CAPath` which can be set by the runtime flag
`-capath=/etc/gaia/cert` for example. It is recommended that the certificate
is kept separate from the main Gaia work folder and in a secure location.

This certificate is used in two places. First, in the communication between the
admin portal and the back-end. Second, by the Vault.

## The Vault

The Vault is a secure storage for secret values like, password, tokens and other
things that the user would like to pass securly into a Pipeline. The Vault is
encrypted using AES cipher technology where the key is derived from the above
certificate and the IV is included in the encrypted content.

The Vault file's location can be configured through the runtime variable called
`VaultPath`. For maximum security it is recommended that this file is kept on an
encrypted, mounted drive. In case there is a breach the drive can be quickly removed
and the file deleted, thus rotating all of the secrets at once, under Gaia.

To create an encrypted MacOSX image follow this guide: [Encrypted Secure Disk Image on Mac](https://www.howtogeek.com/183826/how-to-create-an-encrypted-file-container-disk-image-on-a-mac/).

To create an encrypted disk on Linux follow this guide: [Encrypted Disk Image on Linux](http://freesoftwaremagazine.com/articles/create_encrypted_disk_image_gnulinux/).

The admin will never see the secure values, not when editing, not when adding and not
when looking at the list of secrets. Only the Key names are displayed at all times.

It's possible to Add, Delete, Update and List secrets in the system.
