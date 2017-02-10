#/usr/bin/python

from sklearn.feature_extraction.text import TfidfVectorizer
from sys import argv
from os import walk,write, mkdir

if len(argv) < 2:
    print "Please provide a corpus dir"
    exit(0)


def read_corpus(corpus_dir):
    corpus = []
    filenames =[]
    for root, dirs, files in walk(corpus_dir):
        for fi in files:
            filenames.append(fi)
            f=open(root+"/"+fi,'r')
            lines=f.readlines()
            f.close()
            if len(lines)>0:
                corpus.append(lines[0])
    return corpus,files

def tfidf(corpus):
    vectorizer = TfidfVectorizer(min_df=1)
    return vectorizer.fit_transform(corpus).toarray()


def create_results(matrix, outdir,filanems):
    mkdir(outdir)
    for i in range(0, len(filenames)):
        f = open(outdir+"/"+filanems[i], 'w')
        for val in matrix[i]:
            f.write("%.5f\t" % val)
        f.write("\n")
        f.close()


input_dir=argv[1].replace("/", "")
corpus,filenames=read_corpus(input_dir)
res = tfidf(corpus)
create_results(res, input_dir+"-tfidf", filenames)
