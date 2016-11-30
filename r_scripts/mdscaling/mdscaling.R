#!/usr/bin/Rscript

args <- commandArgs(trailingOnly=TRUE)

if (length(args) < 2 ) { 
			cat("Please provide the similarities csv and the number of dimensions.\n") 
	quit()
}

library(MASS)
s <- as.matrix(read.csv(args[1],header=FALSE))
k <- as.numeric(args[2])
MAXITERATIONS <- 4000
TOLERANCE <- 1e-5
# turn the similarity matrix into a distance matrix
d <- 1/s-1 
fit <- isoMDS(d, k=k, trace=FALSE, tol = TOLERANCE, maxit = MAXITERATIONS)
#fit <- cmdscale(d, eig=FALSE, k=2)

cat(fit$stress)
cat("\n")
for(i in 1:nrow(fit$points)) {
		for(j in 1:ncol(fit$points)) {
				cat(fit$points[i,j])
				cat(" ")
		}
		cat("\n")
}

#for(i in 1:k) {
#		cat(max(fit$points[,i])-min(fit$points[,i]))
#		cat("\n")
#}

