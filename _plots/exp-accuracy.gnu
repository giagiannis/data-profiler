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

set terminal postscript eps size 7,5.0 enhanced color font 'Arial,34'
set output outputFile
set xlabel "Sampling Rate"
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

set title "MSE vs Sampling Rate".titleComment
set ylabel "MSE"
set output outputFile
plot inputFile u 1:2 w lp t system("basename ".inputFile) ls 1
system("epstopdf --outfile=".outputFile."-mse.pdf ".outputFile." && rm ".outputFile)


set title "MAPE vs Sampling Rate".titleComment
set ylabel "MAPE"
set output outputFile
plot inputFile u 1:8 w lp t system("basename ".inputFile) ls 1
system("epstopdf --outfile=".outputFile."-mape.pdf ".outputFile." && rm ".outputFile)

set title "MAPE-log vs Sampling Rate".titleComment
set ylabel "MAPE-log"
set output outputFile
plot inputFile u 1:14 w lp t system("basename ".inputFile) ls 1
system("epstopdf --outfile=".outputFile."-mape-log.pdf ".outputFile." && rm ".outputFile)
