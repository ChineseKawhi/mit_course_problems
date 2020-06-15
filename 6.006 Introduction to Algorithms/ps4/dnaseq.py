#!/usr/bin/env python2.7

import unittest
from dnaseqlib import *

### Utility classes ###

# Maps integer keys to a set of arbitrary values.
class Multidict:
    # Initializes a new multi-value dictionary, and adds any key-value
    # 2-tuples in the iterable sequence pairs to the data structure.
    def __init__(self, pairs=[]):
        self.dic = {}
        for pair in pairs:
            if(self.dic.has_key(pair[0])):
                self.dic[pair[0]].append(pair[1])
            else:
                self.dic[pair[0]] = [pair[1]]
    # Associates the value v with the key k.
    def put(self, k, v):
        if(self.dic.has_key(k)):
            self.dic[k].append(v)
        else:
            self.dic[k] = [v]
    # Gets any values that have been associated with the key k; or, if
    # none have been, returns an empty sequence.
    def get(self, k):
        if(self.dic.has_key(k)):
            return self.dic[k]
        else:
            return []

# Given a sequence of nucleotides, return all k-length subsequences
# and their hashes.  (What else do you need to know about each
# subsequence?)
def subsequenceHashes(seq, k):
    try:
        subseq = ''
        idx = 0
        while(len(subseq) < k):
            subseq += seq.next()
            idx += 1
        rh = RollingHash(subseq)
        while True:
            yield (subseq, rh.current_hash(), idx-k)
            next_item = seq.next()
            idx += 1
            rh.slide(subseq[0], next_item)
            subseq = subseq[1:] + next_item
    except StopIteration:
        return
    

# Similar to subsequenceHashes(), but returns one k-length subsequence
# every m nucleotides.  (This will be useful when you try to use two
# whole data files.)
def intervalSubsequenceHashes(seq, k, m):
    try:
        subseq = ''
        idx = 0
        while(len(subseq) < k):
            subseq += seq.next()
            idx += 1
        rh = RollingHash(subseq)
        while True:
            if((idx-k)%m == 0):
                yield (subseq, rh.current_hash(), idx-k)
            next_item = seq.next()
            idx += 1
            rh.slide(subseq[0], next_item)
            subseq = subseq[1:] + next_item
    except StopIteration:
        return

# Searches for commonalities between sequences a and b by comparing
# subsequences of length k.  The sequences a and b should be iterators
# that return nucleotides.  The table is built by computing one hash
# every m nucleotides (for m >= k).
def getExactSubmatches(a, b, k, m):
    dic = Multidict()
    for s_a, h_a, i_a in intervalSubsequenceHashes(a,k,m):
        # print(s_a)
        dic.put(h_a,(s_a,i_a))
    for s_b, h_b, i_b in intervalSubsequenceHashes(b,k,m):
        # print(s_b)
        it = dic.get(h_b)
        if(len(it) > 0):
            for item in it:
                if(item[0] == s_b):
                    yield (item[1], i_b)
    return

if __name__ == '__main__':
    if len(sys.argv) != 4:
        print 'Usage: {0} [file_a.fa] [file_b.fa] [output.png]'.format(sys.argv[0])
        sys.exit(1)

    # The arguments are, in order: 1) Your getExactSubmatches
    # function, 2) the filename to which the image should be written,
    # 3) a tuple giving the width and height of the image, 4) the
    # filename of sequence A, 5) the filename of sequence B, 6) k, the
    # subsequence size, and 7) m, the sampling interval for sequence
    # A.
    compareSequences(getExactSubmatches, sys.argv[3], (500,500), sys.argv[1], sys.argv[2], 8, 100)
