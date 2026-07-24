# wakeonlan Tool

Ein einfaches Tool, um ein Gerät über Wake-on-LAN (WOL) zu starten.

# Argumente:
  <MAC-Adresse>    Ziel-MAC-Adresse im Format XX:XX:XX:XX:XX:XX \
  (Optional) [Broadcast-IP-Adresse] Standard ist 255.255.255.255

## Beispiele:
```
  ./wakeonlan 00:1A:2B:3C:4D:5E
  ./wakeonlan 00-1A-2B-3C-4D-5E 192.168.0.255
```
