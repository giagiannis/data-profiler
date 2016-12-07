#!/usr/bin/Rscript
# This script utilizes the GenSA package in order to identify the global minimum
# for the stress function. This way,

args <- commandArgs(trailingOnly=TRUE)
if (length(args) < 1 ) { 
    cat("Please provide the points along with their distances\n") 
	quit()
}
makePlot <- FALSE
for(i in 1:length(args)) if(args[i] == "plot") makePlot <- TRUE
library(GenSA)
dataset <- read.csv(args[1])
coords <- as.matrix(dataset[,1:ncol(dataset)-1])
distances <- as.matrix(dataset[ncol(dataset)])

stress <- function(x) {
	y <- numeric(nrow(coords))
	for(i in 1:nrow(coords))  {
			y[i] <- sum(sqrt(sum((x-coords[i,])^2)) - distances[i])
	}
	sqrt(sum(y^2)/sum(distances^2))
}

lower <- numeric(ncol(coords))
for(i in 1:ncol(coords)) lower[i] <- min(coords)-max(distances)
upper <- numeric(ncol(coords))
for(i in 1:ncol(coords)) upper[i] <- max(coords)+max(distances)

sa <- GenSA(fn=stress, lower=lower, upper=upper)
cat(sa$par, sep=" ")
cat("\n")
cat(sa$value)

if(makePlot){
#pdf(paste0(args[1],as.numeric(Sys.time()), ".pdf"))
	pdf(paste0(args[1], ".pdf"))
	plot(coords, main=args[1])
	theta <- seq(0,2*pi, length=200)
	for(i in 1:nrow(coords)) {
		lines(
			  x = distances[i]*cos(theta)+coords[i,1], 
			  y=distances[i]*sin(theta)+coords[i,2])
	}
	grid(NULL,NULL, col="black")
	points(x=sa$par[1], y=sa$par[2], pch=7)
}

