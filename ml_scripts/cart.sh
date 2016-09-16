#!/bin/bash 
# script used to conduct a simple classification technique

TRAINFILE=$1
TESTFILE=$2
FORMULA=""
CLASS_INDEX=0
for i in $(head  -n 1 $TRAINFILE | tr ',' '\n' | grep --invert-match "class"); do
	FORMULA="$FORMULA$i+"
	let CLASS_INDEX=CLASS_INDEX+1
done
let CLASS_INDEX=CLASS_INDEX+1
FORMULA="class ~ ${FORMULA::-1}"
R --slave --no-save << EOF

library(rpart)
train.data <- read.csv("$TRAINFILE")
test.data <- read.csv("$TESTFILE")
formula <- $FORMULA
fit <- rpart(formula, train.data, method="class")
m<-cbind(test.data[,$CLASS_INDEX], predict(fit, test.data, type="class"))
m<-m[,1]-m[,2]
cat(sum(m!=0)/length(test.data[,$CLASS_INDEX]))
EOF
