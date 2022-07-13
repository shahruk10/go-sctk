# Go Wrapper for SCTK

[![CI](https://github.com/shahruk10/go-sctk/actions/workflows/ci.yml/badge.svg)](https://github.com/shahruk10/go-sctk/actions/workflows/ci.yml)

- [SCTK](https://github.com/usnistgov/SCTK) is a toolkit made by NIST that can be used for evaluating the output of automatic speech recognition systems (ASR). It can be used to:

  - Calculate Word Error Rate (WER) and Character Error Rate (CER)

  - Analyze different types of errors made by ASR systems: Substitutions, Insertions and Deletions.

  - Generate alignments between multiple sets of transcripts (reference and hypotheses)

  - Use statistical tests to evaluate the significance in performance delta between ASR systems.

- This repository offers a single binary with a command line interface (CLI) that wraps around the all different tools in SCTK; the CLI has a simple and easy to use interface.

- This repo is a work-in-progress.

---

## Usage Examples

### Evaluating CER

- This example evaluates the character error rate (CER) between reference transcripts and hypothesis transcript generated by a ASR system. The example uses Bengali text but SCTK supports most languages since it expects text with UTF-8 encoding.

```sh
# Creating dummy reference transcript file in CSV format.
cat << EOF > reference.csv
utterance_id,transcript
spk01-utt01,এর মূল্য বার্ষিক দশ লক্ষ ইউরো।
spk02-utt02,খেলাটি চার টেস্ট সিরিজের চূড়ান্ত ছিল।
EOF

# Creating dummy hypothesis transcript from an ASR system, in CSV format.
cat << EOF > hypothesis.csv
utterance_id,transcript
spk01-utt01,এর মূল্য বার্ দশ লক ইউর।
spk02-utt02,খেলা ছার টেস্ট শিরিজের চূড়ান্ত ছিল।
EOF

# Getting the sctk CLI tool from this repository and giving it executable permissions.
version=v0.1.0
wget -O sctk https://github.com/shahruk10/go-sctk/releases/download/${version}/sctk
chmod +x sctk

# Using sctk CLI to evaluate CER and check errors.
#
# Setting `--ignore-first=true` to ignore header row.
# Check `sctk score --help` for documentation of each argument.
sctk score \
  --ignore-first=true \
  --delimiter="," \
  --col-id=0 \
  --col-trn=1 \
  --case-sensitive=false \
  --normalize-unicode=true \
  --cer=true \
  --out=./cer \
  --ref=reference.csv \
  --hyp=hypothesis.csv
```

- Now we can check generated reports in the `./cer` directory.

```
  cer/
  ├── hyp1.trn
  ├── hyp1.trn.dtl
  ├── hyp1.trn.pra
  ├── hyp1.trn.raw
  ├── hyp1.trn.sgml
  ├── hyp1.trn.sys
  └── ref.trn
```

- The `*.sys` file contains a table showing a breakdown of the different types of errors.

  - The results are aggregated for each speaker; `Corr`, `Sub`, `Del` and `Ins` stands for
    the percentage of characters that were correctly decoded, substituted, deleted and inserted
    in the hypothesis respectively.

  ```
                   SYSTEM SUMMARY PERCENTAGES by SPEAKER                      

     ,----------------------------------------------------------------.
     |                              hyp1                              |
     |----------------------------------------------------------------|
     | SPKR   | # Snt # Chr | Corr    Sub    Del    Ins    Err  S.Err |
     |--------+-------------+-----------------------------------------|
     | spk01  |    1     25 | 76.0    0.0   24.0    0.0   24.0  100.0 |
     |--------+-------------+-----------------------------------------|
     | spk02  |    1     33 | 87.9    6.1    6.1    0.0   12.1  100.0 |
     |================================================================|
     | Sum/Avg|    2     58 | 82.8    3.4   13.8    0.0   17.2  100.0 |
     |================================================================|
     |  Mean  |  1.0   29.0 | 81.9    3.0   15.0    0.0   18.1  100.0 |
     |  S.D.  |  0.0    5.7 |  8.4    4.3   12.7    0.0    8.4    0.0 |
     | Median |  1.0   29.0 | 81.9    3.0   15.0    0.0   18.1  100.0 |
     `----------------------------------------------------------------'
  ```

- The `*.dtl` file shows further details of each type of error. This can reveal systematic
  errors and patterns in how the ASR system is transcribing the audio.

  ```
  ... (other useful stuff)

  CONFUSION PAIRS                  Total                 (2)
                                   With >=  1 occurrences (2)
     1:    1  ->  চ ==> ছ
     2:    1  ->  স ==> শ
       -------
           2

  ... (other useful stuff)

  DELETIONS                        Total                 (6)
                                   With >=  1 occurrences (6)
     1:    2  ->  ষ
     2:    2  ->  ি
     3:    1  ->  ক
     4:    1  ->  ট
     5:    1  ->  ো
     6:    1  ->  ্
       -------
           8
  ```

- The `*.pra` file shows alignments between the reference and hypothesis text, which
  makes it easy to see errors in context.

---

## License

[Apache License 2.0](LICENSE)