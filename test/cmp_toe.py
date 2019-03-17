#!/usr/bin/env python3

# Compare different spam mail filters

import json
import os
import shutil
import stat
import subprocess
import sys
import time

toe = './toe.py'
repeats = 5

filters = [
        ( './gonzofilter', 'hamspam.db'),
        ( './test/bogo.sh', '~/.bogofilter'),
        ( './test/qsf.sh', '~/.qsfdb'),
        ( './test/bsfilter.sh', '~/.bsfilter'),
        ( './test/spamprobe.sh', '~/.spamprobe'),
        ]

def mk_name(cmd):
    s = cmd
    i = s.rfind('/')
    if i == -1:
        i = 0
    else:
        i += 1
    s = s[i:]
    j = s.rfind('.')
    if j == -1:
        j = len(s)
    return s[:j]

def clean_db(file_or_dirP):
    file_or_dir = os.path.expanduser(file_or_dirP)
    try:
        s = os.stat(file_or_dir)
        if stat.S_ISDIR(s.st_mode):
            shutil.rmtree(file_or_dir)
        else:
            os.unlink(file_or_dir)
    except FileNotFoundError:
        pass

def max_dict(xs, ys):
    for k, v in xs.items():
        if k in ys:
            if v > ys[k]:
                ys[k] = v
        else:
            ys[k] = v

def min_dict(xs, ys):
    for k, v in xs.items():
        if k in ys:
            if v < ys[k]:
                ys[k] = v
        else:
            ys[k] = v


def main():
    rs = {}
    for cmd, _ in filters:
        rs[cmd] = { 'min': {}, 'max': {} }
    for i in range(repeats):
        for cmd, db in filters:
            clean_db(db)
            start = time.time()
            s = subprocess.check_output([toe, '--json', '--cmd', cmd])
            secs = time.time() - start
            d = json.loads(s)
            d['time_s'] = secs
            max_dict(d, rs[cmd]['max'])
            min_dict(d, rs[cmd]['min'])
    for cmd, d in sorted(rs.items(), key=lambda x:x[0]):
        for m, e in sorted(d.items(), key=lambda x:x[0]):
            print('{:<20}'.format('command'), end= '')
            for k, v in sorted(e.items(), key=lambda x:x[0]):
                print(',{:>9}'.format(k), end='')
            print()
            break
        break
    for cmd, d in sorted(rs.items(), key=lambda x:x[0]):
        for m, e in sorted(d.items(), key=lambda x:x[0]):
            print('{:<20}'.format(mk_name(cmd)+' ('+m+')'), end='')
            for k, v in sorted(e.items(), key=lambda x:x[0]):
                if type(v) == float:
                    print(',{:>9.2f}'.format(v), end='')
                else:
                    print(',{:>9}'.format(v), end='')
            print()


if __name__ == '__main__':
    sys.exit(main())


