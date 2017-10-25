# Asterisk-TTS
Text to speech for Asterisk written in go (re-implementation of http://zaf.github.io/asterisk-googletts/ )

Used in https://github.com/mcilley/asterisk-zabbix-phone-escalation

## Text to speech is invoked via extensions.conf
ex. playing back a greeting when recieving a call by calling out compiled binary placed in agi scripts directory

```
exten => s,n,agi(gtts.agi,"Hello, this is a test message!")

```
