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
pass-through mode (cf. the `-pass` option).

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
                  golang-x-sys-devel \
                  golang-x-text-devel


### Go Modules

Or with more recent Go versions that support Go modules, it's just:

    $ go build -mod=readonly
    $ go test  -mod=readonly -v

Depending on your system you might need to modify the `go.mod`
file, e.g. change the `replace` directive or remove it
completely.

The `-mod=readonly` switch disables automatic changes of the
`go.mod` file in Go versions less than 1.16. (In Go 1.16
this behavior is the default.)

To make sure that locally available dependencies aren't attempted
to be fetched over the net one can set the `GOPROXY` environment
variable to `off`.

On Go version less than 1.17, module support can be disabled by
either setting the `GO111MODULE` environment variable to `off` or
by setting it to `auto` and removing the `go.mod` file.


### Sandboxing

For the sandboxing feature (`-sandbox`) it also requires
[github.com/seccomp/libseccomp-golang][scg] greater than version
0.9.1. Sandboxing support is disabled by default, to enabled it
build with:


    $ GOPATH=$HOME/go:/usr/share/gocode go build -tags sandbox

Tested on:

- Fedora 29 to 33 (compile and execute)
- CentOS 7 (execute, the kernel/libseccomp is too old for the sandbox
  support, though)

## Maildrop

[Maildrop][maildrop] is a fine and actively maintained mail delivery
agent (MDA) that also supports piping messages through external
filters such as Gonzofilter.

Since maildrop doesn't support delivery decisions to be based on
the exit status of an external filter executable we have to call
Gonzofilter in pass-through mode and check the added `X-gonzo:`
header in maildrop.

Example `.mailfilter` snippet:

```
# extra copies for debugging purposes
cc md/copy

xfilter "/usr/local/bin/gonzofilter -pass"

if  ((/^X-gonzo: spam/:H)
{
    to md/spamfilter
}

# catch-all default destination
to maildir
```

Notes:

- This requires maildrop >= 3 (because of the `:H` option)
- maildrop executes the external executable with CWD=$HOME of the
  MDA user - thus, Gonzofilter expects a usable database to
  exist in `$HOME/hamspam.db`. See also the `-db` option to use
  another database location and `toe.py` for how to create such a
  database.

## Security & Reliability

Piping all incoming mail through an executable for spam filtering
makes this executable an interesting and worthwhile target for
remote attacks.

The lexing and parsing required for spam filtering arguably is
much more involved than what is required for mail transport and
delivery. Thus, the added attack surface isn't small nor trivial.

Since Gonzofilter is implemented in Go which provides memory
safety features such as bounds checking, a whole class of bugs is
eliminated from the start. Of course, one can program bugs in
every programming language, but being able to rely on memory
safety features gives you an edge, security wise.

Otherwise, Gonzofilter contains some unit tests, was tested with
a wide range of nasty mail, and it is dogfooded by its author.

In addition, as a [defence in depth][did] measure, Gonzofilter
optionally supports sandboxing under Linux with [seccomp][sc],
e.g.:

    gonzofilter -passthrough -sandbox

### SELinux

This repository also contains an SELinux policy module for
gonzofilter in the `selinux` subdirectory. It can be activated
with the following steps:

    make -f /usr/share/selinux/devel/Makefile gonzofilter.pp
    semodule -i gonzofilter.pp

In comparison with the seccomp sandboxing, SELinux allows more
fine-grained control over file accesses. For example, it's clear
that gonzofilter needs to open some files, read some and
read/write some others. Thus the involved syscalls need to be
allowed.  This is also what the SELinux policy does, but it does
so while restricting those accesses to files that are labeled
with specific labels. Meaning that the gonzofilter process can
write to the hamspam database and `/tmp` but not to any other
location.

Although coming up with a minimal white-list of syscalls is kind
of tedious, implementing the sandbox approach is arguably more
straight forward than creating a SELinux policy module. At least
the SELinux learning process is more involved.


### Further Considerations

When using a mail filter that is written in a memory unsafe
language (such as C), one has to ask herself how well it is
reviewed and tested for security issues. Perhaps it got some
auditing by other developers and it was fuzzed - perhaps not -
even if it's packaged by Linux distributions.

For example, Bogofilter, implemented in C, was started in 2002 or
so, is packaged by some Linux Distributions and has a good
classification and runtime performance. However, it's a little
frightening that a bit of fuzzing in 2019 easily finds a row of
memory safety issues: [out-of-bounds reads #118][bf118] and
[#126][bf126], [memory leaks #119][bf119] and [#125][bf125],
[buffer management issues #120] and [#121][bf121],
[heap-buffer-overflows/out-of-bounds writes #122][bf122],
[#123][bf123] and [#124][bf124]. Likely meaning that in
the preceding years nobody cared to fuzz it. (Or perhaps somebody
fuzzed it but not publicized the findings.) Depending on in what
shape the code base is and how much maintenance manpower is
available, it may take some time for found issues to be fixed (3
months for the above examples, not verified).

On the other hand, Bogofilter even had a history of heap-buffer
out-of-bounds writes in the years 2004 to 2012: [documented in 5
CVEs][bfcve] ([see also][bfcve2]). And still the reviews and
fixes that resulted from those findings left some low hanging
fuzzing fruit, years later.

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
[maildrop]: https://www.courier-mta.org/maildropfilter.html

[bf118]: https://sourceforge.net/p/bogofilter/bugs/118/
[bf119]: https://sourceforge.net/p/bogofilter/bugs/119/
[bf120]: https://sourceforge.net/p/bogofilter/bugs/120/
[bf121]: https://sourceforge.net/p/bogofilter/bugs/121/
[bf122]: https://sourceforge.net/p/bogofilter/bugs/122/
[bf123]: https://sourceforge.net/p/bogofilter/bugs/123/
[bf124]: https://sourceforge.net/p/bogofilter/bugs/124/
[bf125]: https://sourceforge.net/p/bogofilter/bugs/125/
[bf126]: https://sourceforge.net/p/bogofilter/bugs/126/
[bfcve]: https://cve.mitre.org/cgi-bin/cvekey.cgi?keyword=bogofilter
[bfcve2]: https://bogofilter.sourceforge.io/security/

[scg]: https://github.com/seccomp/libseccomp-golang
[sc]: https://en.wikipedia.org/wiki/Seccomp
[did]: https://en.wikipedia.org/wiki/Defense_in_depth_(computing)
