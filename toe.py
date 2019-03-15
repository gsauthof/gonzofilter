#!/usr/bin/env python3

# Train-on-error and test a spamfilter
# Expects disjoint train/test message sets in a directory
# structure like this:
#
# ex
# ├── learn
# │   ├── ham
# │   └── spam
# └── test
#     ├── ham
#     └── spam
#
# It's recommended to randomly split a set of ham/spam messages
# (e.g. with shuf) into the directories.
# For example, when having 1000 messages each one can use a 400/600
# split between the learn and test directories.
#
# 2019, Georg Sauthoff

import os
import random
import subprocess
import sys

base = 'ex'

cmd = './gonzofilter'
#cmd = './bogo.sh'


def learn(filename, spam=False):
    action = '-spam' if spam else '-ham'
    p = subprocess.check_output([cmd, action, '-in', filename])

def classifies(filename, spam=False):
    p = subprocess.run([cmd, '-check', '-in', filename], stdout=subprocess.DEVNULL)
    x = 11 if spam else 10
    if p.returncode not in (10, 11):
        raise RuntimeError(f'Unexpected exit status for {filename}: {p.returncode})')
    return p.returncode == x

def toe():
    d = base + '/learn/ham/'
    hams = [ d + x for x in os.listdir(d) ]
    random.shuffle(hams)
    d = base + '/learn/spam/'
    spams = [ d + x for x in os.listdir(d) ]
    random.shuffle(spams)

    ham_msg_cnt = 0
    spam_msg_cnt = 0

    for i in range(50):
        ham_msg_cnt += 1
        learn(hams.pop())
        spam_msg_cnt += 1
        learn(spams.pop(), spam=True)

    while True:
        rest_ham = []
        rest_spam = []
        old_ham_msg_cnt, old_spam_msg_cnt = ham_msg_cnt, spam_msg_cnt
        while hams or spams:
            if spams:
                x = spams.pop()
                if classifies(x, spam=True):
                    rest_spam.append(x)
                else:
                    spam_msg_cnt += 1
                    learn(x, spam=True)
                    if hams:
                        x = hams.pop()
                        ham_msg_cnt += 1
                        learn(x)
                        continue
            if hams:
                x = hams.pop()
                if classifies(x):
                    rest_ham.append(x)
                else:
                    ham_msg_cnt += 1
                    learn(x)
                    if spams:
                        x = spams.pop()
                        spam_msg_cnt += 1
                        learn(x, spam=True)
                        continue
        if rest_ham:
            hams += rest_ham
        if rest_spam:
            spams += rest_spam
        if not hams and not spams:
            break
        if ham_msg_cnt == old_ham_msg_cnt and spam_msg_cnt == old_spam_msg_cnt:
            break
    print(f'Learned {ham_msg_cnt} ham messages and {spam_msg_cnt} spam messages')

def test_class(spam=False):
    klasse = 'spam' if spam else 'ham'
    d = base + f'/test/{klasse}/'
    xs = [ d + x for x in os.listdir(d) ]
    true, false = 0, 0
    for x in xs:
        if classifies(x, spam):
            true += 1
        else:
            false += 1
    return true, false

def test():
    true_positives, false_negatives = test_class(spam=True)
    true_negatives, false_positives = test_class()
    print(f'True (Spam) Positives: {true_positives}, False Negatives: {false_negatives}, True negatives: {true_negatives}, False Positives: {false_positives}')
    sensitivity = true_positives / (true_positives + false_negatives)
    specificity = true_negatives / (true_negatives + false_positives)
    accuracy = (true_positives + true_negatives) / (true_positives + false_negatives + true_negatives + false_positives)
    print(f'Spam detection Sensitivity: {sensitivity}, Specificity: {specificity}, Accuracy: {accuracy}')

def main():
    toe()
    test()

if __name__ == '__main__':
    sys.exit(main())


