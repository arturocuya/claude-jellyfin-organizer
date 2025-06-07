hello! i'm arturo, a movie and tv show enthusiast. i have a vast collection of media that i have scanned from my own legally owned dvd's and blu-rays.

i need your help organizing them on my jellyfin media library

i have media files at this path: "{{.InputPath}}"

to organize my files, here's what you should do:

1. examine the files to determine what media they contain (movie or tv show). you can determine the type also in the next step.
2. find the exact name of the media on imdb, so that you can get the imdb id
3. consider the documentation of how to organize jellyfin media. i'll attach it
4. use the available tools to copy and rename my files and place them in the right folder

the folder for my jellyfin movies is {{.MoviesFolder}}, and the one for my jellyfin shows is {{.ShowsFolder}}

now here's the documentation on how to organize a jellyfin media library

{{.JellyfinDocs}}

having read that, please prefer using the imdb id on the file names to ensure proper metadata download!

feel free to add the imdb id suffix to existing folders if they need it.

you should return a plan of what you want to do before executing the tools to copy to my jellyfin library. wait for my confirmation to do the final copy.

thanks!