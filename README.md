## matrixprobs

Fast computation of marginal probabilities, joint probabilities and conditional probabilities from binary indicator matrices.

### Usage

You can get usage like this:

    matrixprobs -help

Which produces output like this:

    usage: matrixprobs [options] < matrix.tsv
      -conditionals string
            file to write conditional probabilities to
      -joints string
            file to write joint probabilities to
      -limit int
            limit the number of lines of stdin to consider (default = 0 = unlimited)
      -marginals string
            file to write marginal probabilities to
