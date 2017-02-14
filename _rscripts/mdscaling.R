#!/usr/bin/Rscript

args <- commandArgs(trailingOnly=TRUE)

if (length(args) < 2 ) { 
			cat("Please provide the similarities csv and the number of dimensions.\n") 
	quit()
}

#library(MASS)
d <- as.matrix(read.csv(args[1],header=FALSE))
k <- as.numeric(args[2])
MAXITERATIONS <- 1000
TOLERANCE <- 1e-4
EPSILON <- 1e-10
#fit <- isoMDS(d, k=k, trace=FALSE, tol = TOLERANCE, maxit = MAXITERATIONS, p=2)
#cat(fit$stress)
#cat("\n")
#for(i in 1:nrow(fit$points)) {
#		for(j in 1:ncol(fit$points)) {
#				cat(fit$points[i,j])
#				cat(" ")
#		}
#		cat("\n")
#}
#
fit <- cmdscale(d, k=k, eig=TRUE)
cat(fit$GOF[2])
cat("\n")
for(i in 1:nrow(fit$points)) {
		for(j in 1:ncol(fit$points)) {
				cat(fit$points[i,j])
				cat(" ")
		}
		cat("\n")
}
