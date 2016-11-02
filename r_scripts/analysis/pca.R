#!/usr/bin/Rscript
# 
# Author: ggian
# Last updated: Tue Oct  4 16:02:00 EEST 2016
# This script analyzes an input dataset, and outputs the eigenvectors
# of the dataset, scaled by their respective variance (as determined by 
# the respective eigenvalue)

args <- commandArgs(trailingOnly=TRUE)
datafile <- args[1]
dataset <- read.csv(datafile)
cols = ncol(dataset)
fit <- princomp(dataset, cor=TRUE)


# prints the number of the eigenvalues/vectors
cat(length(fit$sd)^2)
cat("\n")

# prints the eigenvalues

variances=c()
sumOfVariances = sum(as.numeric(fit$sd)^2)
for (i in 1:(cols)) {
	variances[i] = (fit$sd[i]^2)/sumOfVariances
}

#print(variances)
# prints the eigenvectors
for (i in 1:(cols*cols)) {
	cat(as.numeric(variances[(i-1)%/%cols + 1])*fit$loadings[i])
	cat(" ")
}
cat("\n")
