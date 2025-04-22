# Proxyprint

## Flow

### Tunneling
Version number: 1
1. Tunneler connects to server
2. Tunneler sends tunnel header (0xFFFFFFFF)
3. Server disconnects on bad header or sends version (4 bytes)
4. Tunneler determines if version is acceptable or disconnects
    - Tunnel should most likely quit entirely if version is disagreeable
5. Tunneler sends password (with 8-byte length prefix, big endian)
6. Server sends OK (0x00000001) or ERROR (0x00000002) bytes
7. When the server has a client connection ready for tunneling, the server sends the ready bytes (0xFEFEFEFE)
    - If error, tunnel should most likely quit entirely (unless retrying with different password)
8. The tunnel having received and verified the bytes sends its ready bytes (0xFDFDFDFD)
9. Communication commences
