#!/usr/bin/gnuplot
# mds-stress visualizes the results of the mds util (stress module)

if(!exists("inputFile") || !exists("outputFile")) {
	print("The following params must be defined: inputFile, outputFile");
	exit(0);
}


if(!exists("titleComment")) {
	titleComment="" 
} else {
	titleComment=" (".titleComment.")"
}

set title "Stress values vs dimensionality".titleComment
set terminal postscript eps size 7,5.0 enhanced color font 'Arial,34'
set output outputFile
set xlabel "Dimensions"
set ylabel "Stress"
#set logscale x
#set xrange[1000:]
#set yrange [0:1]
#set logscale x
set grid
set key top right

set style line 1 lt 1 pt 7 lc rgb "#000000" lw 4 ps 4
set style line 2 lt 1 pt 2 lc rgb "#333333" lw 4 ps 4
set style line 3 lt 1 pt 4 lc rgb "#777777" lw 4 ps 4
set style line 4 lt 1 pt 10 lc rgb "#999999" lw 4 ps 4
set style line 5 lt 1 pt 13 lc rgb "#aaaaaa" lw 4 ps 4

plot for [i=1:words(inputFile)] word(inputFile,i) u 1:2 w lp t system("basename ".word(inputFile,i)) ls i
system("epstopdf ".outputFile." && rm ".outputFile)
