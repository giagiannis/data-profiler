#!/usr/bin/Rscript
args <- commandArgs(trailingOnly=TRUE)
datafile <- args[1]
dataset <- read.csv(datafile)
cols = ncol(dataset)
fit <- princomp(dataset, cor=TRUE)


# prints the number of the eigenvalues/vectors
cat(length(fit$sd)^2)
cat("\n")

# prints the eigenvalues
#for (i in 1:(cols)) {
#	cat(as.numeric(fit$sd[i]^2))
#	cat(" ")
#}
#cat("\n")

# prints the eigenvectors
for (i in 1:(cols*cols)) {
	v <- as.numeric(fit$sd[(i-1)%/%cols + 1]^2)*fit$loadings[i]
	cat(v)
	cat(" ")
}
cat("\n")
