Media Sorter
------------------

This purpose of this tool is to search for media files (audio, video, photo) in a given directory and move them based on the camera model, and the date the photo was taken.

It uses the exif data provided by ExifTool to gather the nessicary info.

It has five arguments:

-sourceDir = The directory containing the images to be moved (trailing slashes are stripped) Default: /tmp
-desDir = The destination you want the images sorted into. (trailing slashes are stripped) Default: /tmp
-info = Read media exif, process it into a struct, and print out the gathered info. then exit without further action.  Currently: filename, source path, destination Path, Make, Model, Camera Serial Number, Date and Time it was taken.
-metadata = Read exif data and print out everything it frinds in the files and exit without further action.
-dry-run = *WIP: DOES NOT WORK YET* This is to just tell you what it would do (fails after directory "creation")


Functionality
--------
1) Walks through the provided sourceDir and a list if files with extentions maching the list in procDir
2) it then iterates over that list of files and process them.

If `-metadata` is passed it will grab the exif info here, print the results and move to next file
If `-info` is passed it will grab the exif info, get the pieces it wants to process and print out the results, then continue to the next file

Once the exif is processed it will determine if the source and destination paths are the same and if so, it will stop there.  *this is to see if previously processed files would get moved*

At this point it will get the permission mode for the lowest directy that currently exists and create the nessicary directories in the path using that mode.

After that it checks if the file already exists and if it does trys to figure out if its the same file or not, otherwise it moves the file.

If the desitionation file already exists:
  It first checks if the sha256 hash of the source an existing destination file match, if so it moves on to the next file without further action
  If they are not the same then it prepends the camera's serial number to the filename and try's again.
If the new desitionation doesn't exist it moves the file and moves one, if it does exists:
  It checks the sha256 hashes again and if they match move on, otherwise it bails on the file with an print out as such.

Known Supported Cameras
-----------
Nikon D5300
Panisonic Lumix GH5
Panisonic Lumix GH5S
Go Pro HERO4 Silver - Requires exif flag -ee - (add support for THM files?)

Partial Support
------------
ZOOM Handy Recorder H6 (auido) -- Model = [Originator] -- Needs Make and Serial Number

Camera/device Support forth comming
---------------
Panasonic Lumix DMC-FZ35
Sony Handy Cam HDR-SR7
Sony Handy Cam HDR-XR520V
HTC Glacier (phone)
myTouch 4G (phone)
Samsung SGH-I717 (phone)
Samsung SM-G900T (phone)
T-Mobile myTouch 3G (phone)


TODO
--------
Can I compare file size for dup files: if size != then not same add serial, else get hashes?
Complete Dry Run flagging (currently fails because the destination dir wasn't created)
Tests
Parallel proccessing
Fix Nikon serial number handling to get an int instead of scientific notation
what if after serial number is added to filename the file already exists and doesn't match? then what
What if there is no seral number to grab? ex Zoom H6
Better Handling of Exif Fields to get info needed when hiding in other fields
Do I need to make a list of files or should I proccess while walking the directory structure?
Proper Logging
Error handling??
  what if a directory couldn't be read
  what if a directory couldn't be made
  what if file couldn't get move?

Future Features
--------------
Add Rsync to copy files from sdCard to a tmp Dir for processing
Add monitoring to watch for sdCards to be plugged in and to see if its from a camera
Add a watch dir for files not from an sdCard
Docker??


Running
--------------
mediaSorter -sourceDir=/mnt/sdCard -destDir=/mnt/visualMedia
