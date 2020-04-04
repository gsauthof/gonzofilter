This repository contains Gonzofilter, a Bayes classifying spam
mail filter written in [Go][go].

2019, Georg Sauthoff <mail@gms.tf>, GPLv3+

## Getting started

Build a new database with some already classified messages
(either manually classified or classified with another classifier):

    $ ./toe.py

See the short `toe.py` script for details.

To classify new messages:

    $ ./gonzofilter -check -in path/to/maildir/msg
    $ echo $?

The exit status 10 stands for a 'ham' classification result while
11 stands for 'spam'.

For integration with a mail-delivery-agent it also supports a
pass-through mode (cf. the `-pass option`).

To just use it as tokenizer:

    $ ./gonzofilter -dump-mark -in path/to/maildir/msg

See also `-h` for additional commands and options.

## Classification Performance

Gonzofilter implements a [naive Bayes classifier][bayes] for
classifying messages into [spam][spam] and ham classes. It's
called naive because some simplifying assumptions are applied,
such as the [independence][ind] of word occurrences. Naive Bayes
classifiers are used for text classification since the 1970ies or
so, partly because they are simple to implement, but also because
they often perform surprisingly well. They were popularized for
filtering Spam in 2002 by Paul Graham's article [A Plan for
Spam][paul].

To evaluate the performance of Gonzofilter I created the small
benchmark script (cf. `test/cmp_toe.py`) that runs a
train-on-error (TOE) procedure (cf. `toe.py`) with Gonzofilter
and some other open-source Bayes spam filters. The following
results are from a Fedora 29 x86_64 system (Intel i7-6600U CPU,
16 GiB RAM), with Gonzofilter compiled with the Fedora packaged
Go and Fedora packaged dependencies, and the other filters also
installed from the Fedora repositories.

    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    command              FN    FP  accuracy   lham   lspam   sensi   speci   time_s
    ――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――
    gonzofilter (max)   115    14      0.93    167     168    0.88    0.99    44.06
    gonzofilter (min)    72     5      0.90    132     133    0.81    0.98    37.74
    bogo (max)          172     7      0.87    218     218    0.75    0.99    50.79
    bogo (min)          149     3      0.85    188     188    0.71    0.99    44.24
    bsfilter (max)      178    16      0.90    212     212    0.83    0.99   341.31
    bsfilter (min)      103     8      0.84    165     164    0.70    0.97   279.01
    qsf (max)           148    43      0.90    227     227    0.86    0.96    89.74
    qsf (min)            83    24      0.86    196     195    0.75    0.93    73.40
    spamprobe (max)      90    15      0.92    153     153    0.87    0.98    77.70
    spamprobe (min)      78    10      0.91    136     136    0.85    0.97    64.89
    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Note that Gonzofilter has the highest [accuracy][accuracy] and fastest
runtime, while only using a moderate number of training messages.

For the experiment I randomly selected 1000 ham mails from my
inbox and I selected the latest 1000 spam mails from my junk mail
box. I then randomly split each message set into learn and test
sets (with a 4 to 6 ratio). The train-on-error (TOE) procedure
then uses the learning sets for training the classifier and the
test sets for checking its performance, i.e. its [sensitivity,
specificity and accuracy][accuracy].

In the above table, FP stands for false-positive which means that
a message is falsely identified as spam, whereas FN means
false-negative, i.e. a message is falsely identified as ham (i.e.
non-spam). The lham and lspam counts are the number of messages
the train-of-error procedure consumed until no further
classification errors occurred in the test set (which has a size
of 400 messages). The runtime denotes the runtime of a `toe.py`
run. Since the train-on-error procedure shuffles each training
set, the performance may vary from run to run. Thus, for each
filter, the TOE run it repeated 5 times and the table contains the
minimum and maximum results.

There is some room for variability when implementing a Bayes
classification model. For example, you can model a message as a
set or as [bag (multi-set)][bag] of words. Gonzofilter uses the
[bag of words][bow] approach, it computes the probability
weights in log space and uses [pseudocounts][pcounts] to deal
with words that didn't occur during learning.
In comparison, the [Bogofilter][bogo] spam filter applies some
[extensions][lj] to the naive Bayes model and uses libgsl for
some statistical computations. [Spamprobe][spamp] documents that
it also uses the frequencies of two word phrases for its
statistical model.

Another important factor for classification performance is how a
message is tokenized into words. Gonzofilter goes to some lengths
to tokenize, decode and normalize a message into a stream of
words.  That means it decodes base64 encoded parts,
quoted-printable parts, understands MIME, ignores non-text
attachments, removes HTML tags and comments (but keeps the
referenced URLs in some tags), translates HTML entities, converts
various character encodings into UTF-8, normalizes some special
code points like [soft-hyphen][shy], but keeps some punctuation
characters attached to words, and more. Also, the words are
prefixed with a location tag, e.g. words from the subject Header
are prefixed with `h:subject:` while body words are prefixed with
`b:`. Last but not least, words smaller than 4 characters and
larger than 32 characters are ignored, as well as some headers
containing ids and dates.

In comparison, the [Bogofilter][bogo] lexer also does some word
prefixing and character conversion. The [Quick Spam Filter
(QSF)][qsf] filter rejects/ignores mails larger than 512 KiB.
[Bsfilter][bsf] also seems to do some character conversion which
may fail with an uncaught exception (observed this with one
message from my test set).

For [bsfilter][bsf], the runtime difference can be explained by
the different implementation languages. Gonzofilter is
implemented in [Go][go], which is natively compiled with garbage
collected memory management. Although [garbage collection][gc]
may be challenging for performance in some scenarios, the
Gonzofilter implementation is careful to avoid buffer churning,
to avoid unnecessarily copying memory around and to tokenize
messages efficiently in general. [Bsfilter][bsf] is implemented
in Ruby, which is compiled to Byte-Code that is interpreted
without a [JIT][jit] by the Ruby VM. [Bogofilter][bogo] and the
[Quick Spam Filter (QSF)][qsf] are implemented in C, where
Bogofilter uses a [Flex][flex] generated tokenizer, while
[Spamprobe][spamp] is implemented in C++.

## Build Instructions

Compile it:

    $ GOPATH=$HOME/go:/usr/share/gocode go build

Run the unittests:

    $ GOPATH=$HOME/go:/usr/share/gocode go test -v

Set the GOPATH differently if the dependencies are installed
elsewhere or you want to use another workspace location.

It only needs a few extra dependencies:

    - go.etcd.io/bbolt
    - golang.org/x/text/encoding
    - golang.org/x/sys/unix

They can be installed with `go get` or the distribution's package
manager. For example, on Fedora:

    # dnf install golang-etcd-bbolt-devel \
                  golang-golangorg-text-devel \
                  golang-github-golang-sys-devel

Tested on:

- Fedora 29, 31

## Motivation

- Have an accessible platform to test different text
  classification approaches
- Evaluate the trade-offs when writing something exposed as a
  mail filter in a memory-safe language
- Learn a new programming language (Go) - which has some
  interesting features, arguably is better designed than Java,
  but also has some shortcomings

[bayes]: https://en.wikipedia.org/wiki/Naive_Bayes_classifier
[spam]: https://en.wikipedia.org/wiki/Email_spam
[paul]: http://www.paulgraham.com/spam.html
[ind]: https://en.wikipedia.org/wiki/Independence_(probability_theory)
[lj]: https://www.linuxjournal.com/article/6467
[qsf]: http://www.ivarch.com/programs/qsf/
[bsf]: http://sourceforge.jp/projects/bsfilter/
[accuracy]: https://en.wikipedia.org/wiki/Sensitivity_and_specificity
[spamp]: http://spamprobe.sourceforge.net/
[bogo]: http://bogofilter.sourceforge.net/
[bag]: https://en.wikipedia.org/wiki/Multiset
[bow]: https://en.wikipedia.org/wiki/Bag-of-words_model
[shy]: https://en.wikipedia.org/wiki/Soft_hyphen
[flex]: https://en.wikipedia.org/wiki/Flex_(lexical_analyser_generator)
[jit]: https://en.wikipedia.org/wiki/Just-in-time_compilation
[gc]: https://en.wikipedia.org/wiki/Garbage_collection_(computer_science)
[go]: https://en.wikipedia.org/wiki/Go_(programming_language)
[pcounts]: https://en.wikipedia.org/wiki/Additive_smoothing
