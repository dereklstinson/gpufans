# gpufans
Fan control for desktop use of Nvidia Tesla (for now and maybe forever) gpus.

Core function works!  It controls the fans like a champ.  It also reads the fan speeds.  Though, it doesn't output fan speeds at the moment.

To Do:

Make a arduino circuit.

Make this a part of systemcl

Make this or another program read the fan speeds. (If this is a part of systemcl. I think I need to make another program.)

Make sure config.json file can be read. (Might have to make a .folder in the $HOME folder for this.)


Requirements:
You will need nvidia-smi.

You will also have to:
```
go get github.com/tarm/serial 
```