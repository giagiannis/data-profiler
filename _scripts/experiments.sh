#!/bin/bash
# script used to execute the experiment


# variables that need to be defined before execution
echo $GOPATH
[ "$OPERATORS" == "" ] && echo "OPERATORS variable needed" && exit 0
[ "$DATASET_PATH" == "" ] && echo "DATASET_PATH variable needed" && exit 0

[ "$MDSSCRIPT" == "" ] && export MDSSCRIPT="$GOPATH/src/github.com/giagiannis/data-profiler/_rscripts/mdscaling.R"
[ "$SVMSCRIPT" == "" ] && export SVMSCRIPT="$GOPATH/src/github.com/giagiannis/data-profiler/_rscripts/svm-appx.R"

THREADS=20
REPETITIONS=20
TREE_SCALE=0.9

b="$(basename $DATASET_PATH)/$TREE_SCALE"
mkdir -p $b 2>/dev/null
BASE="$b/base"
mkdir -p $BASE 2>/dev/null
RESULTS_DIR="$b/experiments"
mkdir -p $RESULTS_DIR 2>/dev/null
PLOTS_DIR="$b/plots"
mkdir -p $PLOTS_DIR 2>/dev/null

function prepare_dataset() {
		LOGFILE="$BASE/log"
		echo "Calculating similarities"
		data-profiler-utils similarities -i $DATASET_PATH -o $BASE/matrix.sim -opt concurrency=$THREADS,tree.scale=$TREE_SCALE -l $LOGFILE
		echo "Calculating MDS"
		data-profiler-utils mds -k 5 -l $LOGFILE -m coords -o $BASE/mds -sc $MDSSCRIPT -sim $BASE/matrix.sim
		echo "Calculating MDS (evaluating goodness of fit)"
		data-profiler-utils mds -k 30 -l $LOGFILE -m gof -o $BASE/mds -sc $MDSSCRIPT -sim $BASE/matrix.sim

		echo "Running operators"
		for op in $OPERATORS; do 
			b="$(basename $op)"
			echo -e "\tOperator $op"
			data-profiler-utils train -i $DATASET_PATH -l $LOGFILE -p $THREADS -o $BASE/$b-scores -s $op -t "nothing"
		done
}

function execute_experiments() {
		SR="$(seq -s, 0.05 0.05 1.0)"
		LOGFILE="$RESULTS_DIR/log"

		echo "Executing accuracy and ordering experiments"
		for op in $OPERATORS; do 
				echo -e "\tOperator $op"
				b="$(basename $op)"
				data-profiler-utils exp-accuracy -c $BASE/mds-coords -i $BASE/matrix.sim.idx -l $LOGFILE -ml $SVMSCRIPT -o $RESULTS_DIR/$b-accuracy.csv -r $REPETITIONS -s $BASE/$b-scores -sr "$SR" -t $THREADS
				data-profiler-utils exp-ordering -c $BASE/mds-coords -i $BASE/matrix.sim.idx -l $LOGFILE -ml $SVMSCRIPT -o $RESULTS_DIR/$b-ordering.csv -r $REPETITIONS -s $BASE/$b-scores -sr "$SR" -t $THREADS
		done

}

function plot_experiments() {
		GNUPLOT_ACCURACY="/home/giannis/data-profiler/_plots/exp-accuracy.gnu"
		GNUPLOT_ORDERING="/home/giannis/data-profiler/_plots/exp-ordering.gnu"

		echo "Generating plots"

		for op in $OPERATORS; do
				echo -e "\tOperator $op"
				b=$(basename $op)
				gnuplot -e "inputFile='$RESULTS_DIR/$b-accuracy.csv';outputFile='$PLOTS_DIR/$b-accuracy'" $GNUPLOT_ACCURACY 
				gnuplot -e "inputFile='$RESULTS_DIR/$b-ordering.csv';outputFile='$PLOTS_DIR/$b-ordering'" $GNUPLOT_ORDERING
		done
}

prepare_dataset
execute_experiments
plot_experiments
