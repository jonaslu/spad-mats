# Setup
This is the sister-repo to a blog-post discussing commit atomicity and literacy.
Reading that post first will make much more sense of the rest in this file.
[link](link).

See the [README.md](README.md) for what the programs do.

I've included the SQL that was ran against the postgres database so you can follow along (or try it on your own repo).

# The overall state
If you've run the steps in the [README.md](README.md) you should now have a database running with commits from the 100 most popular repos sampled down to 1000 commits per repo.

Only the default branch (master or main) will be considered for commits. The reason being that including all branches may skew the result as a branch can be work in progress. The main or master branch should reflect how commits are handled when they are ready for public consumption.

In the list there are some repos that are not mostly code (an example is the awesome-* lists). There were 24 such repos at the time of the test. Since the atomicity and literacy mostly concerns understanding code you could argue that they should be omitted and that they will obscure the result.

On the other hand you could argue that they represent the state of repos regardless of what they contain and even a change in an .md file could be properly documented. To end the discussion I did omit the repos that are not mostly code and ran the stats once again - but the result barely moved when they were removed. Thus I've included these in the result since the discussion holds with (or without) them.

The database contains two tables: log and message. The log table contains the sha, date, author, added lines, removed lines, filename and what gitrepo it belongs to.

The message table contains the sha of the message, the text of the body of the message and a word count of the message.

A commit can hardly be literate if it doesn't have a commit body. This means that the commit has some text message after the subject line which is the first line. The subject line hasn't been included for the simple reason that making sense of the entire commit in 60 characters is very slim and breaks the etiquette that the subject line should summarize the longer message in the body.

Now let's dive in and see how many if we can measure the atomicity, literacy and correlation of the two. We then end with a discussion of the results.

# Atomicity of the commits
Turning first to the atomicity of a commit - I've defined the commit "size" as rows added + rows removed since a removal should be scrutinized just as hard as an addition when you're changing a program.

Using a [histogram](https://en.wikipedia.org/wiki/Histogram) we can now plot how many commits of a certain size that fall into different buckets. Starting with all commits in the entire database regardless of what repo it belongs to divided into 20 buckets with the smallest commit as a lower bound and the largest commit size as the upper bound:

```SQL
with commit_size as (
  select sum(added) + sum(removed) as amount_change from log_filtered group by sha
), max_commit_size as (
  select max(commit_size.amount_change) as max from commit_size
), total_commits as (
  select count(*) as total_commits from commit_size
)
select width_bucket(commit_size.amount_change, 1, max_commit_size.max, 19) as bucket,
int4range(min(commit_size.amount_change)::integer, max(commit_size.amount_change)::integer, '[]') as range,
count(*) as freq,
round((count(*)::decimal/(select total_commits from total_commits))*100, 4) as percent_of_commits
from commit_size, max_commit_size
group by bucket
order by bucket desc;
```

Result:
```
 bucket |      range      | freq  | percent_of_commits
--------+-----------------+-------+--------------------
     20 | [845705,845706) |     1 |             0.0011
     17 | [730768,737592) |     2 |             0.0023
     12 | [516446,516447) |     1 |             0.0011
     11 | [483132,483133) |     1 |             0.0011
      9 | [380774,392270) |     2 |             0.0023
      7 | [271175,271176) |     1 |             0.0011
      6 | [231892,266153) |     3 |             0.0034
      5 | [181357,206133) |     5 |             0.0057
      4 | [144490,175846) |     9 |             0.0102
      3 | [89340,119835)  |    14 |             0.0158
      2 | [44666,88213)   |    32 |             0.0362
      1 | [1,43693)       | 87759 |            99.2244
      0 | [0,1)           |   615 |             0.6953
(13 rows)
```

99% percent of commits fall within the range 1-43693 lines changed. Not so interesting as the whole result is obscured by some very very large commits.

Let's filter it down to commits in the range 1-2000 lines of code into a 100 bucket histogram to get some more insight on the 99% range above.

```SQL
with commit_size as (
  select sum(added) + sum(removed) as amount_change, sha from log_filtered group by sha
), total_commits as (
  select count(*) as total_commits from commit_size where amount_change < 2000
)
select width_bucket(commit_size.amount_change, 1, 2000, 99) as bucket,
int4range(min(commit_size.amount_change)::integer, max(commit_size.amount_change)::integer, '[]') as range,
count(*) as freq,
round((count(*)::decimal/(select total_commits from total_commits))*100, 4) as percent_of_commits
from commit_size, total_commits
where commit_size.amount_change < 2000
group by bucket
order by bucket desc;
```

Result:
```
 bucket |    range    | freq  | percent_of_commits
--------+-------------+-------+--------------------
     99 | [1981,2000) |     8 |             0.0092
     98 | [1960,1979) |    17 |             0.0195
     97 | [1940,1957) |    13 |             0.0149
     96 | [1920,1939) |     8 |             0.0092
     95 | [1902,1919) |    13 |             0.0149
     94 | [1879,1897) |    10 |             0.0115
     93 | [1859,1878) |    14 |             0.0160
     92 | [1841,1859) |     6 |             0.0069
     91 | [1821,1835) |    10 |             0.0115
     90 | [1800,1818) |    12 |             0.0137
     89 | [1779,1798) |    15 |             0.0172
     88 | [1759,1776) |     7 |             0.0080
     87 | [1741,1758) |    12 |             0.0137
     86 | [1718,1734) |    10 |             0.0115
     85 | [1698,1717) |     9 |             0.0103
     84 | [1677,1697) |    18 |             0.0206
     83 | [1659,1676) |    11 |             0.0126
     82 | [1638,1655) |    20 |             0.0229
     81 | [1620,1637) |    19 |             0.0218
     80 | [1597,1613) |    16 |             0.0183
     79 | [1576,1596) |    14 |             0.0160
     78 | [1561,1573) |     7 |             0.0080
     77 | [1536,1553) |    12 |             0.0137
     76 | [1517,1536) |    16 |             0.0183
     75 | [1497,1513) |    13 |             0.0149
     74 | [1479,1494) |     9 |             0.0103
     73 | [1456,1476) |    20 |             0.0229
     72 | [1435,1453) |    12 |             0.0137
     71 | [1415,1435) |    19 |             0.0218
     70 | [1396,1414) |    17 |             0.0195
     69 | [1375,1395) |    18 |             0.0206
     68 | [1355,1374) |    21 |             0.0240
     67 | [1337,1353) |    26 |             0.0298
     66 | [1315,1334) |    16 |             0.0183
     65 | [1295,1313) |    12 |             0.0137
     64 | [1274,1294) |    14 |             0.0160
     63 | [1253,1273) |    24 |             0.0275
     62 | [1233,1253) |    21 |             0.0240
     61 | [1214,1233) |    19 |             0.0218
     60 | [1193,1212) |    16 |             0.0183
     59 | [1173,1193) |    31 |             0.0355
     58 | [1152,1170) |    21 |             0.0240
     57 | [1132,1152) |    31 |             0.0355
     56 | [1112,1132) |    32 |             0.0366
     55 | [1093,1111) |    19 |             0.0218
     54 | [1072,1092) |    29 |             0.0332
     53 | [1051,1071) |    34 |             0.0389
     52 | [1031,1051) |    33 |             0.0378
     51 | [1011,1031) |    31 |             0.0355
     50 | [991,1011)  |    40 |             0.0458
     49 | [971,990)   |    38 |             0.0435
     48 | [951,971)   |    41 |             0.0469
     47 | [930,951)   |    42 |             0.0481
     46 | [910,930)   |    49 |             0.0561
     45 | [890,910)   |    60 |             0.0687
     44 | [870,890)   |    48 |             0.0550
     43 | [850,870)   |    42 |             0.0481
     42 | [829,850)   |    51 |             0.0584
     41 | [809,829)   |    53 |             0.0607
     40 | [789,809)   |    68 |             0.0779
     39 | [769,789)   |    68 |             0.0779
     38 | [749,769)   |    58 |             0.0664
     37 | [728,749)   |    78 |             0.0893
     36 | [708,728)   |    57 |             0.0653
     35 | [688,708)   |    81 |             0.0928
     34 | [668,688)   |    80 |             0.0916
     33 | [648,668)   |    89 |             0.1019
     32 | [627,648)   |    93 |             0.1065
     31 | [607,627)   |    83 |             0.0950
     30 | [587,607)   |    89 |             0.1019
     29 | [567,587)   |   100 |             0.1145
     28 | [547,567)   |   102 |             0.1168
     27 | [526,547)   |   124 |             0.1420
     26 | [506,526)   |   130 |             0.1489
     25 | [486,506)   |   137 |             0.1569
     24 | [466,486)   |   144 |             0.1649
     23 | [446,466)   |   157 |             0.1798
     22 | [426,446)   |   156 |             0.1786
     21 | [405,426)   |   187 |             0.2141
     20 | [385,405)   |   180 |             0.2061
     19 | [365,385)   |   215 |             0.2462
     18 | [345,365)   |   250 |             0.2863
     17 | [325,345)   |   318 |             0.3641
     16 | [304,325)   |   310 |             0.3550
     15 | [284,304)   |   380 |             0.4351
     14 | [264,284)   |   367 |             0.4202
     13 | [244,264)   |   457 |             0.5233
     12 | [224,244)   |   526 |             0.6023
     11 | [203,224)   |   621 |             0.7111
     10 | [183,203)   |   710 |             0.8130
      9 | [163,183)   |   848 |             0.9710
      8 | [143,163)   |  1003 |             1.1485
      7 | [123,143)   |  1296 |             1.4840
      6 | [102,123)   |  1777 |             2.0348
      5 | [82,102)    |  2244 |             2.5695
      4 | [62,82)     |  3252 |             3.7238
      3 | [42,62)     |  5043 |             5.7746
      2 | [22,42)     |  9202 |            10.5369
      1 | [1,22)      | 54337 |            62.2196
      0 | [0,1)       |   615 |             0.7042
(100 rows)
```

As we can see the amount of commits between 1 and 22 lines are 62% of the total amount of commits.

What does the atomicity look within each repo? Let's plot a histogram of 4 buckets within the range 1-80 as some semi-scientific cutoff for how atomic a commit is.

```SQL
with commit_size as (
  select sum(added) + sum(removed) as amount_change, gitrepo from log group by sha, gitrepo
), message_max_min as (
  select count(*) as total_commits,
         gitrepo
    from commit_size group by gitrepo
)
select width_bucket(c.amount_change, 1, 80, 2) as bucket,
int4range(min(c.amount_change)::integer, max(c.amount_change)::integer, '[]') as range,
count(*) as freq,
c.gitrepo,
round((count(*)::decimal/(select total_commits from message_max_min e where e.gitrepo = c.gitrepo))*100, 2) as precent_of_commits
from commit_size c
group by c.gitrepo, bucket
order by c.gitrepo, bucket desc;
```

Result:
```
 bucket |    range     | freq |                    gitrepo                    | precent_of_commits
--------+--------------+------+-----------------------------------------------+--------------------
      3 | [80,18892)   |  170 | /30-seconds/30-seconds-of-code                |              17.02
      2 | [41,80)      |   92 | /30-seconds/30-seconds-of-code                |               9.21
      1 | [1,41)       |  726 | /30-seconds/30-seconds-of-code                |              72.67
      0 | [0,1)        |   11 | /30-seconds/30-seconds-of-code                |               1.10
      3 | [90,3651)    |   32 | /996icu/996.ICU                               |               3.21
      2 | [41,77)      |   35 | /996icu/996.ICU                               |               3.51
      1 | [1,41)       |  875 | /996icu/996.ICU                               |              87.68
      0 | [0,1)        |   56 | /996icu/996.ICU                               |               5.61
      3 | [81,19416)   |  133 | /adam-p/markdown-here                         |              17.97
      2 | [41,80)      |   79 | /adam-p/markdown-here                         |              10.68
      1 | [1,41)       |  519 | /adam-p/markdown-here                         |              70.14
      0 | [0,1)        |    9 | /adam-p/markdown-here                         |               1.22
      3 | [80,2798)    |   26 | /airbnb/javascript                            |               2.60
      2 | [42,79)      |   37 | /airbnb/javascript                            |               3.70
      1 | [1,40)       |  936 | /airbnb/javascript                            |              93.60
      0 | [0,1)        |    1 | /airbnb/javascript                            |               0.10
      3 | [80,42160)   |  325 | /angular/angular                              |              32.50
      2 | [41,80)      |  127 | /angular/angular                              |              12.70
      1 | [1,41)       |  547 | /angular/angular                              |              54.70
      0 | [0,1)        |    1 | /angular/angular                              |               0.10
      3 | [80,26468)   |  185 | /angular/angular.js                           |              18.50
      2 | [41,80)      |  108 | /angular/angular.js                           |              10.80
      1 | [1,41)       |  704 | /angular/angular.js                           |              70.40
      0 | [0,1)        |    3 | /angular/angular.js                           |               0.30
      3 | [80,737592)  |  227 | /ansible/ansible                              |              22.72
      2 | [41,80)      |   88 | /ansible/ansible                              |               8.81
      1 | [1,41)       |  678 | /ansible/ansible                              |              67.87
      0 | [0,1)        |    6 | /ansible/ansible                              |               0.60
      3 | [80,266153)  |  143 | /ant-design/ant-design                        |              14.30
      2 | [41,80)      |   72 | /ant-design/ant-design                        |               7.20
      1 | [1,41)       |  785 | /ant-design/ant-design                        |              78.50
      3 | [80,170703)  |  344 | /apache/incubator-echarts                     |              34.40
      2 | [41,79)      |  105 | /apache/incubator-echarts                     |              10.50
      1 | [1,41)       |  548 | /apache/incubator-echarts                     |              54.80
      0 | [0,1)        |    3 | /apache/incubator-echarts                     |               0.30
      3 | [80,8599)    |  277 | /apple/swift                                  |              27.70
      2 | [41,80)      |  141 | /apple/swift                                  |              14.10
      1 | [1,41)       |  580 | /apple/swift                                  |              58.00
      0 | [0,1)        |    2 | /apple/swift                                  |               0.20
      3 | [83,13683)   |   84 | /atom/atom                                    |               8.41
      2 | [41,80)      |   66 | /atom/atom                                    |               6.61
      1 | [1,41)       |  843 | /atom/atom                                    |              84.38
      0 | [0,1)        |    6 | /atom/atom                                    |               0.60
      3 | [100,619)    |    5 | /avelino/awesome-go                           |               0.50
      2 | [41,77)      |    8 | /avelino/awesome-go                           |               0.80
      1 | [1,41)       |  987 | /avelino/awesome-go                           |              98.70
      3 | [80,80656)   |  160 | /bitcoin/bitcoin                              |              16.00
      2 | [41,80)      |  121 | /bitcoin/bitcoin                              |              12.10
      1 | [1,41)       |  715 | /bitcoin/bitcoin                              |              71.50
      0 | [0,1)        |    4 | /bitcoin/bitcoin                              |               0.40
      3 | [80,34839)   |  269 | /chartjs/Chart.js                             |              26.90
      2 | [41,80)      |  106 | /chartjs/Chart.js                             |              10.60
      1 | [1,41)       |  623 | /chartjs/Chart.js                             |              62.30
      0 | [0,1)        |    2 | /chartjs/Chart.js                             |               0.20
      3 | [81,90010)   |  111 | /chrislgarry/Apollo-11                        |              25.52
      2 | [41,80)      |   51 | /chrislgarry/Apollo-11                        |              11.72
      1 | [1,41)       |  262 | /chrislgarry/Apollo-11                        |              60.23
      0 | [0,1)        |   11 | /chrislgarry/Apollo-11                        |               2.53
      3 | [80,49087)   |  177 | /CyC2018/CS-Notes                             |              17.74
      2 | [41,79)      |   80 | /CyC2018/CS-Notes                             |               8.02
      1 | [1,41)       |  716 | /CyC2018/CS-Notes                             |              71.74
      0 | [0,1)        |   25 | /CyC2018/CS-Notes                             |               2.51
      3 | [80,37117)   |  261 | /d3/d3                                        |              26.10
      2 | [41,80)      |  134 | /d3/d3                                        |              13.40
      1 | [1,41)       |  602 | /d3/d3                                        |              60.20
      0 | [0,1)        |    3 | /d3/d3                                        |               0.30
      3 | [175,486)    |    3 | /danistefanovic/build-your-own-x              |               0.94
      1 | [1,34)       |  315 | /danistefanovic/build-your-own-x              |              98.75
      0 | [0,1)        |    1 | /danistefanovic/build-your-own-x              |               0.31
      3 | [80,21926)   |  363 | /denoland/deno                                |              36.30
      2 | [41,80)      |  138 | /denoland/deno                                |              13.80
      1 | [1,41)       |  498 | /denoland/deno                                |              49.80
      0 | [0,1)        |    1 | /denoland/deno                                |               0.10
      3 | [80,68287)   |  169 | /django/django                                |              16.93
      2 | [41,80)      |  105 | /django/django                                |              10.52
      1 | [1,41)       |  719 | /django/django                                |              72.04
      0 | [0,1)        |    5 | /django/django                                |               0.50
      3 | [90,2925)    |   28 | /donnemartin/system-design-primer             |               9.09
      2 | [41,71)      |   14 | /donnemartin/system-design-primer             |               4.55
      1 | [1,41)       |  249 | /donnemartin/system-design-primer             |              80.84
      0 | [0,1)        |   17 | /donnemartin/system-design-primer             |               5.52
      3 | [80,2227)    |   63 | /doocs/advanced-java                          |              13.85
      2 | [42,80)      |   33 | /doocs/advanced-java                          |               7.25
      1 | [1,41)       |  322 | /doocs/advanced-java                          |              70.77
      0 | [0,1)        |   37 | /doocs/advanced-java                          |               8.13
      3 | [84,1733)    |   19 | /EbookFoundation/free-programming-books       |               1.90
      2 | [41,79)      |   18 | /EbookFoundation/free-programming-books       |               1.80
      1 | [1,41)       |  961 | /EbookFoundation/free-programming-books       |              96.10
      0 | [0,1)        |    2 | /EbookFoundation/free-programming-books       |               0.20
      3 | [80,23577)   |  364 | /elastic/elasticsearch                        |              36.40
      2 | [42,80)      |  120 | /elastic/elasticsearch                        |              12.00
      1 | [1,41)       |  516 | /elastic/elasticsearch                        |              51.60
      3 | [80,25477)   |  121 | /electron/electron                            |              12.12
      2 | [41,80)      |  105 | /electron/electron                            |              10.52
      1 | [1,41)       |  770 | /electron/electron                            |              77.15
      0 | [0,1)        |    2 | /electron/electron                            |               0.20
      3 | [80,19485)   |  188 | /ElemeFE/element                              |              18.80
      2 | [41,80)      |  103 | /ElemeFE/element                              |              10.30
      1 | [1,41)       |  709 | /ElemeFE/element                              |              70.90
      3 | [82,3596)    |   91 | /expressjs/express                            |               9.10
      2 | [41,79)      |   91 | /expressjs/express                            |               9.10
      1 | [1,41)       |  817 | /expressjs/express                            |              81.70
      0 | [0,1)        |    1 | /expressjs/express                            |               0.10
      3 | [84,3072)    |  113 | /facebook/create-react-app                    |              11.30
      2 | [41,80)      |   95 | /facebook/create-react-app                    |               9.50
      1 | [1,41)       |  789 | /facebook/create-react-app                    |              78.90
      0 | [0,1)        |    3 | /facebook/create-react-app                    |               0.30
      3 | [81,63790)   |  279 | /facebook/react                               |              27.90
      2 | [41,80)      |  115 | /facebook/react                               |              11.50
      1 | [1,41)       |  599 | /facebook/react                               |              59.90
      0 | [0,1)        |    7 | /facebook/react                               |               0.70
      3 | [80,65001)   |  250 | /facebook/react-native                        |              25.00
      2 | [41,80)      |  108 | /facebook/react-native                        |              10.80
      1 | [1,41)       |  639 | /facebook/react-native                        |              63.90
      0 | [0,1)        |    3 | /facebook/react-native                        |               0.30
      3 | [80,32719)   |  322 | /flutter/flutter                              |              32.23
      2 | [41,80)      |  135 | /flutter/flutter                              |              13.51
      1 | [1,41)       |  538 | /flutter/flutter                              |              53.85
      0 | [0,1)        |    4 | /flutter/flutter                              |               0.40
      3 | [108,380775) |   44 | /FortAwesome/Font-Awesome                     |              61.97
      2 | [46,80)      |    3 | /FortAwesome/Font-Awesome                     |               4.23
      1 | [1,30)       |   24 | /FortAwesome/Font-Awesome                     |              33.80
      3 | [80,173271)  |  122 | /freeCodeCamp/freeCodeCamp                    |              12.27
      2 | [41,80)      |   80 | /freeCodeCamp/freeCodeCamp                    |               8.05
      1 | [1,41)       |  790 | /freeCodeCamp/freeCodeCamp                    |              79.48
      0 | [0,1)        |    2 | /freeCodeCamp/freeCodeCamp                    |               0.20
      3 | [80,16599)   |  170 | /gatsbyjs/gatsby                              |              17.03
      2 | [41,77)      |  109 | /gatsbyjs/gatsby                              |              10.92
      1 | [1,41)       |  716 | /gatsbyjs/gatsby                              |              71.74
      0 | [0,1)        |    3 | /gatsbyjs/gatsby                              |               0.30
      3 | [80,8765)    |  178 | /getify/You-Dont-Know-JS                      |              17.80
      2 | [41,80)      |  112 | /getify/You-Dont-Know-JS                      |              11.20
      1 | [1,41)       |  695 | /getify/You-Dont-Know-JS                      |              69.50
      0 | [0,1)        |   15 | /getify/You-Dont-Know-JS                      |               1.50
      3 | [82,212)     |    9 | /github/gitignore                             |               0.90
      2 | [41,78)      |   21 | /github/gitignore                             |               2.10
      1 | [1,41)       |  954 | /github/gitignore                             |              95.40
      0 | [0,1)        |   16 | /github/gitignore                             |               1.60
      3 | [81,23632)   |  184 | /gohugoio/hugo                                |              18.40
      2 | [41,80)      |  101 | /gohugoio/hugo                                |              10.10
      1 | [1,41)       |  708 | /gohugoio/hugo                                |              70.80
      0 | [0,1)        |    7 | /gohugoio/hugo                                |               0.70
      3 | [80,193264)  |  304 | /golang/go                                    |              30.40
      2 | [41,80)      |  144 | /golang/go                                    |              14.40
      1 | [1,41)       |  550 | /golang/go                                    |              55.00
      0 | [0,1)        |    2 | /golang/go                                    |               0.20
      3 | [80,14321)   |   73 | /goldbergyoni/nodebestpractices               |               7.31
      2 | [41,80)      |   96 | /goldbergyoni/nodebestpractices               |               9.62
      1 | [1,41)       |  802 | /goldbergyoni/nodebestpractices               |              80.36
      0 | [0,1)        |   27 | /goldbergyoni/nodebestpractices               |               2.71
      3 | [80,1947)    |   18 | /GoThinkster/realworld                        |               4.21
      2 | [41,79)      |   16 | /GoThinkster/realworld                        |               3.74
      1 | [1,41)       |  370 | /GoThinkster/realworld                        |              86.45
      0 | [0,1)        |   24 | /GoThinkster/realworld                        |               5.61
      3 | [83,8585)    |  105 | /h5bp/Front-end-Developer-Interview-Questions |              19.06
      2 | [42,80)      |   33 | /h5bp/Front-end-Developer-Interview-Questions |               5.99
      1 | [1,41)       |  405 | /h5bp/Front-end-Developer-Interview-Questions |              73.50
      0 | [0,1)        |    8 | /h5bp/Front-end-Developer-Interview-Questions |               1.45
      3 | [80,25132)   |  137 | /h5bp/html5-boilerplate                       |              13.70
      2 | [41,80)      |   66 | /h5bp/html5-boilerplate                       |               6.60
      1 | [1,41)       |  780 | /h5bp/html5-boilerplate                       |              78.00
      0 | [0,1)        |   17 | /h5bp/html5-boilerplate                       |               1.70
      3 | [82,2032)    |  135 | /hakimel/reveal.js                            |              13.50
      2 | [41,79)      |  108 | /hakimel/reveal.js                            |              10.80
      1 | [1,41)       |  757 | /hakimel/reveal.js                            |              75.70
      3 | [80,3566)    |  128 | /httpie/httpie                                |              12.80
      2 | [41,80)      |   82 | /httpie/httpie                                |               8.20
      1 | [1,41)       |  784 | /httpie/httpie                                |              78.40
      0 | [0,1)        |    6 | /httpie/httpie                                |               0.60
      3 | [81,16993)   |  314 | /iluwatar/java-design-patterns                |              31.46
      2 | [41,80)      |   97 | /iluwatar/java-design-patterns                |               9.72
      1 | [1,41)       |  567 | /iluwatar/java-design-patterns                |              56.81
      0 | [0,1)        |   20 | /iluwatar/java-design-patterns                |               2.00
      3 | [81,999)     |   73 | /jlevy/the-art-of-command-line                |               9.08
      2 | [41,80)      |   76 | /jlevy/the-art-of-command-line                |               9.45
      1 | [1,41)       |  654 | /jlevy/the-art-of-command-line                |              81.34
      0 | [0,1)        |    1 | /jlevy/the-art-of-command-line                |               0.12
      3 | [80,805)     |   14 | /josephmisiti/awesome-machine-learning        |               1.67
      2 | [42,76)      |   14 | /josephmisiti/awesome-machine-learning        |               1.67
      1 | [1,41)       |  810 | /josephmisiti/awesome-machine-learning        |              96.43
      0 | [0,1)        |    2 | /josephmisiti/awesome-machine-learning        |               0.24
      3 | [80,11949)   |  109 | /jquery/jquery                                |              10.90
      2 | [41,80)      |   96 | /jquery/jquery                                |               9.60
      1 | [1,41)       |  793 | /jquery/jquery                                |              79.30
      0 | [0,1)        |    2 | /jquery/jquery                                |               0.20
      3 | [100,1459)   |   13 | /justjavac/free-programming-books-zh_CN       |               2.11
      2 | [46,72)      |    5 | /justjavac/free-programming-books-zh_CN       |               0.81
      1 | [1,38)       |  595 | /justjavac/free-programming-books-zh_CN       |              96.75
      0 | [0,1)        |    2 | /justjavac/free-programming-books-zh_CN       |               0.33
      3 | [80,3991)    |  160 | /jwasham/coding-interview-university          |              16.00
      2 | [41,80)      |   91 | /jwasham/coding-interview-university          |               9.10
      1 | [1,41)       |  737 | /jwasham/coding-interview-university          |              73.70
      0 | [0,1)        |   12 | /jwasham/coding-interview-university          |               1.20
      3 | [89,51277)   |   45 | /kamranahmedse/developer-roadmap              |              14.85
      2 | [52,77)      |    5 | /kamranahmedse/developer-roadmap              |               1.65
      1 | [1,37)       |  221 | /kamranahmedse/developer-roadmap              |              72.94
      0 | [0,1)        |   32 | /kamranahmedse/developer-roadmap              |              10.56
      3 | [80,9723)    |   68 | /kdn251/interviews                            |              17.13
      2 | [41,80)      |   89 | /kdn251/interviews                            |              22.42
      1 | [1,41)       |  227 | /kdn251/interviews                            |              57.18
      0 | [0,1)        |   13 | /kdn251/interviews                            |               3.27
      3 | [256,257)    |    1 | /kelseyhightower/nocode                       |              25.00
      2 | [43,44)      |    1 | /kelseyhightower/nocode                       |              25.00
      1 | [1,13)       |    2 | /kelseyhightower/nocode                       |              50.00
      3 | [80,206133)  |  331 | /kubernetes/kubernetes                        |              33.10
      2 | [41,80)      |  102 | /kubernetes/kubernetes                        |              10.20
      1 | [1,41)       |  567 | /kubernetes/kubernetes                        |              56.70
      3 | [80,14601)   |   75 | /laravel/laravel                              |               7.50
      2 | [41,79)      |   59 | /laravel/laravel                              |               5.90
      1 | [1,41)       |  863 | /laravel/laravel                              |              86.30
      0 | [0,1)        |    3 | /laravel/laravel                              |               0.30
      3 | [80,37072)   |  205 | /lodash/lodash                                |              20.50
      2 | [41,80)      |  114 | /lodash/lodash                                |              11.40
      1 | [1,41)       |  678 | /lodash/lodash                                |              67.80
      0 | [0,1)        |    3 | /lodash/lodash                                |               0.30
      3 | [80,112897)  |  297 | /Microsoft/PowerToys                          |              29.73
      2 | [41,78)      |   85 | /Microsoft/PowerToys                          |               8.51
      1 | [1,41)       |  602 | /Microsoft/PowerToys                          |              60.26
      0 | [0,1)        |   15 | /Microsoft/PowerToys                          |               1.50
      3 | [80,483133)  |  354 | /Microsoft/terminal                           |              35.40
      2 | [41,80)      |  123 | /Microsoft/terminal                           |              12.30
      1 | [1,41)       |  518 | /Microsoft/terminal                           |              51.80
      0 | [0,1)        |    5 | /Microsoft/terminal                           |               0.50
      3 | [80,271176)  |  342 | /Microsoft/TypeScript                         |              34.20
      2 | [41,80)      |  146 | /Microsoft/TypeScript                         |              14.60
      1 | [1,41)       |  510 | /Microsoft/TypeScript                         |              51.00
      0 | [0,1)        |    2 | /Microsoft/TypeScript                         |               0.20
      3 | [80,24051)   |  200 | /Microsoft/vscode                             |              20.00
      2 | [41,80)      |  127 | /Microsoft/vscode                             |              12.70
      1 | [1,41)       |  673 | /Microsoft/vscode                             |              67.30
      3 | [81,6151)    |   68 | /MisterBooo/LeetCodeAnimation                 |              39.31
      2 | [42,78)      |   19 | /MisterBooo/LeetCodeAnimation                 |              10.98
      1 | [1,31)       |   82 | /MisterBooo/LeetCodeAnimation                 |              47.40
      0 | [0,1)        |    4 | /MisterBooo/LeetCodeAnimation                 |               2.31
      3 | [80,38135)   |  222 | /moby/moby                                    |              22.20
      2 | [41,80)      |  116 | /moby/moby                                    |              11.60
      1 | [1,41)       |  660 | /moby/moby                                    |              66.00
      0 | [0,1)        |    2 | /moby/moby                                    |               0.20
      3 | [80,198752)  |  201 | /moment/moment                                |              20.10
      2 | [41,80)      |   91 | /moment/moment                                |               9.10
      1 | [1,41)       |  707 | /moment/moment                                |              70.70
      0 | [0,1)        |    1 | /moment/moment                                |               0.10
      3 | [80,150165)  |  266 | /mrdoob/three.js                              |              26.63
      2 | [41,80)      |   96 | /mrdoob/three.js                              |               9.61
      1 | [1,41)       |  637 | /mrdoob/three.js                              |              63.76
      3 | [80,49294)   |  269 | /mui-org/material-ui                          |              26.90
      2 | [41,80)      |  139 | /mui-org/material-ui                          |              13.90
      1 | [1,41)       |  592 | /mui-org/material-ui                          |              59.20
      3 | [81,13275)   |  173 | /netdata/netdata                              |              20.28
      2 | [41,80)      |   68 | /netdata/netdata                              |               7.97
      1 | [1,41)       |  608 | /netdata/netdata                              |              71.28
      0 | [0,1)        |    4 | /netdata/netdata                              |               0.47
      3 | [80,265303)  |  218 | /nodejs/node                                  |              21.80
      2 | [41,80)      |  129 | /nodejs/node                                  |              12.90
      1 | [1,41)       |  650 | /nodejs/node                                  |              65.00
      0 | [0,1)        |    3 | /nodejs/node                                  |               0.30
      3 | [80,2659)    |   94 | /nvbn/thefuck                                 |               9.40
      2 | [41,80)      |  138 | /nvbn/thefuck                                 |              13.80
      1 | [1,41)       |  768 | /nvbn/thefuck                                 |              76.80
      3 | [80,1395)    |   77 | /ohmyzsh/ohmyzsh                              |               7.70
      2 | [41,80)      |   68 | /ohmyzsh/ohmyzsh                              |               6.80
      1 | [1,41)       |  849 | /ohmyzsh/ohmyzsh                              |              84.90
      0 | [0,1)        |    6 | /ohmyzsh/ohmyzsh                              |               0.60
      3 | [86,35762)   |   34 | /ossu/computer-science                        |               4.64
      2 | [41,78)      |   29 | /ossu/computer-science                        |               3.96
      1 | [1,41)       |  666 | /ossu/computer-science                        |              90.86
      0 | [0,1)        |    4 | /ossu/computer-science                        |               0.55
      3 | [81,7252)    |  121 | /pallets/flask                                |              12.10
      2 | [42,80)      |  102 | /pallets/flask                                |              10.20
      1 | [1,41)       |  774 | /pallets/flask                                |              77.40
      0 | [0,1)        |    3 | /pallets/flask                                |               0.30
      3 | [80,61323)   |  167 | /PanJiaChen/vue-element-admin                 |              16.70
      2 | [41,80)      |   78 | /PanJiaChen/vue-element-admin                 |               7.80
      1 | [1,41)       |  750 | /PanJiaChen/vue-element-admin                 |              75.00
      0 | [0,1)        |    5 | /PanJiaChen/vue-element-admin                 |               0.50
      3 | [80,52454)   |  351 | /pytorch/pytorch                              |              35.10
      2 | [41,80)      |  153 | /pytorch/pytorch                              |              15.30
      1 | [1,40)       |  496 | /pytorch/pytorch                              |              49.60
      3 | [81,5016)    |  101 | /rails/rails                                  |              10.10
      2 | [41,80)      |  101 | /rails/rails                                  |              10.10
      1 | [1,41)       |  796 | /rails/rails                                  |              79.60
      0 | [0,1)        |    2 | /rails/rails                                  |               0.20
      3 | [80,845706)  |  408 | /ReactiveX/RxJava                             |              40.80
      2 | [41,80)      |  118 | /ReactiveX/RxJava                             |              11.80
      1 | [1,41)       |  474 | /ReactiveX/RxJava                             |              47.40
      3 | [81,3338)    |  113 | /redis/redis                                  |              11.30
      2 | [41,80)      |  117 | /redis/redis                                  |              11.70
      1 | [1,41)       |  770 | /redis/redis                                  |              77.00
      3 | [82,181358)  |  136 | /reduxjs/redux                                |              13.60
      2 | [41,79)      |   56 | /reduxjs/redux                                |               5.60
      1 | [1,41)       |  800 | /reduxjs/redux                                |              80.00
      0 | [0,1)        |    8 | /reduxjs/redux                                |               0.80
      3 | [82,1393)    |   15 | /resume/resume.github.com                     |               8.43
      2 | [48,77)      |   11 | /resume/resume.github.com                     |               6.18
      1 | [1,35)       |  149 | /resume/resume.github.com                     |              83.71
      0 | [0,1)        |    3 | /resume/resume.github.com                     |               1.69
      3 | [80,6269)    |  249 | /rust-lang/rust                               |              24.90
      2 | [41,80)      |  126 | /rust-lang/rust                               |              12.60
      1 | [1,41)       |  622 | /rust-lang/rust                               |              62.20
      0 | [0,1)        |    3 | /rust-lang/rust                               |               0.30
      3 | [80,2233)    |   28 | /ryanmcdermott/clean-code-javascript          |               7.49
      2 | [42,79)      |   48 | /ryanmcdermott/clean-code-javascript          |              12.83
      1 | [1,41)       |  298 | /ryanmcdermott/clean-code-javascript          |              79.68
      3 | [80,392270)  |  262 | /Semantic-Org/Semantic-UI                     |              26.20
      2 | [41,80)      |   71 | /Semantic-Org/Semantic-UI                     |               7.10
      1 | [1,41)       |  665 | /Semantic-Org/Semantic-UI                     |              66.50
      0 | [0,1)        |    2 | /Semantic-Org/Semantic-UI                     |               0.20
      3 | [80,15936)   |  266 | /shadowsocks/shadowsocks-windows              |              26.60
      2 | [41,80)      |  123 | /shadowsocks/shadowsocks-windows              |              12.30
      1 | [1,41)       |  598 | /shadowsocks/shadowsocks-windows              |              59.80
      0 | [0,1)        |   13 | /shadowsocks/shadowsocks-windows              |               1.30
      3 | [80,843)     |    6 | /sindresorhus/awesome                         |               0.75
      2 | [41,73)      |   16 | /sindresorhus/awesome                         |               1.99
      1 | [1,41)       |  780 | /sindresorhus/awesome                         |              97.14
      0 | [0,1)        |    1 | /sindresorhus/awesome                         |               0.12
      3 | [80,4953)    |  151 | /Snailclimb/JavaGuide                         |              15.13
      2 | [41,80)      |   62 | /Snailclimb/JavaGuide                         |               6.21
      1 | [1,41)       |  762 | /Snailclimb/JavaGuide                         |              76.35
      0 | [0,1)        |   23 | /Snailclimb/JavaGuide                         |               2.30
      3 | [80,18071)   |  135 | /socketio/socket.io                           |              13.50
      2 | [41,80)      |   78 | /socketio/socket.io                           |               7.80
      1 | [1,40)       |  784 | /socketio/socket.io                           |              78.40
      0 | [0,1)        |    3 | /socketio/socket.io                           |               0.30
      3 | [80,58447)   |  204 | /spring-projects/spring-boot                  |              20.40
      2 | [41,80)      |   81 | /spring-projects/spring-boot                  |               8.10
      1 | [1,41)       |  713 | /spring-projects/spring-boot                  |              71.30
      0 | [0,1)        |    2 | /spring-projects/spring-boot                  |               0.20
      3 | [80,46677)   |  183 | /storybooks/storybook                         |              18.30
      2 | [41,80)      |   76 | /storybooks/storybook                         |               7.60
      1 | [1,41)       |  740 | /storybooks/storybook                         |              74.00
      0 | [0,1)        |    1 | /storybooks/storybook                         |               0.10
      3 | [82,730769)  |  242 | /tensorflow/models                            |              24.22
      2 | [41,80)      |   74 | /tensorflow/models                            |               7.41
      1 | [1,41)       |  677 | /tensorflow/models                            |              67.77
      0 | [0,1)        |    6 | /tensorflow/models                            |               0.60
      3 | [80,88213)   |  305 | /tensorflow/tensorflow                        |              30.50
      2 | [41,80)      |  131 | /tensorflow/tensorflow                        |              13.10
      1 | [1,41)       |  563 | /tensorflow/tensorflow                        |              56.30
      0 | [0,1)        |    1 | /tensorflow/tensorflow                        |               0.10
      3 | [80,45375)   |  249 | /TheAlgorithms/Python                         |              25.00
      2 | [41,80)      |  211 | /TheAlgorithms/Python                         |              21.18
      1 | [1,41)       |  518 | /TheAlgorithms/Python                         |              52.01
      0 | [0,1)        |   18 | /TheAlgorithms/Python                         |               1.81
      3 | [81,169949)  |  151 | /tonsky/FiraCode                              |              33.93
      2 | [41,80)      |   48 | /tonsky/FiraCode                              |              10.79
      1 | [1,41)       |  233 | /tonsky/FiraCode                              |              52.36
      0 | [0,1)        |   13 | /tonsky/FiraCode                              |               2.92
      3 | [80,5208)    |  177 | /torvalds/linux                               |              17.70
      2 | [41,80)      |  112 | /torvalds/linux                               |              11.20
      1 | [1,41)       |  711 | /torvalds/linux                               |              71.10
      3 | [80,12763)   |  240 | /trekhleb/javascript-algorithms               |              26.82
      2 | [42,80)      |   91 | /trekhleb/javascript-algorithms               |              10.17
      1 | [1,41)       |  559 | /trekhleb/javascript-algorithms               |              62.46
      0 | [0,1)        |    5 | /trekhleb/javascript-algorithms               |               0.56
      3 | [80,11343)   |  155 | /twbs/bootstrap                               |              15.50
      2 | [41,80)      |   85 | /twbs/bootstrap                               |               8.50
      1 | [1,41)       |  754 | /twbs/bootstrap                               |              75.40
      0 | [0,1)        |    6 | /twbs/bootstrap                               |               0.60
      3 | [83,24617)   |   82 | /typicode/json-server                         |              10.83
      2 | [41,78)      |   36 | /typicode/json-server                         |               4.76
      1 | [1,41)       |  637 | /typicode/json-server                         |              84.15
      0 | [0,1)        |    2 | /typicode/json-server                         |               0.26
      3 | [85,1676)    |   14 | /vinta/awesome-python                         |               1.40
      2 | [44,79)      |    9 | /vinta/awesome-python                         |               0.90
      1 | [1,41)       |  975 | /vinta/awesome-python                         |              97.60
      0 | [0,1)        |    1 | /vinta/awesome-python                         |               0.10
      3 | [271,11771)  |    4 | /vuejs/awesome-vue                            |               0.40
      2 | [44,45)      |    1 | /vuejs/awesome-vue                            |               0.10
      1 | [1,39)       |  995 | /vuejs/awesome-vue                            |              99.50
      3 | [80,32222)   |  193 | /vuejs/vue                                    |              19.30
      2 | [41,80)      |  144 | /vuejs/vue                                    |              14.40
      1 | [1,41)       |  659 | /vuejs/vue                                    |              65.90
      0 | [0,1)        |    4 | /vuejs/vue                                    |               0.40
      3 | [80,21149)   |  208 | /webpack/webpack                              |              20.80
      2 | [41,80)      |  112 | /webpack/webpack                              |              11.20
      1 | [1,41)       |  679 | /webpack/webpack                              |              67.90
      0 | [0,1)        |    1 | /webpack/webpack                              |               0.10
      3 | [80,2826)    |  213 | /zeit/next.js                                 |              21.30
      2 | [41,80)      |  114 | /zeit/next.js                                 |              11.40
      1 | [1,41)       |  673 | /zeit/next.js                                 |              67.30
(380 rows)
```

Having examined the atomicity of the commits in the sampled repos let's turn to the literacy.

# Missing bodies
The first interesting question is how many commits out of the sampled repos are entirely missing a body? It's hard to be a literate commit if it doesn't have any text attached to it.

``` SQL
with total_commits as (
  select count(distinct sha) as shas from log_filtered
)
select round((count(distinct l.sha)::decimal/(select shas from total_commits))*100, 2) as percent_empty_commits
from log_filtered l left join message m on l.sha = m.sha where m.sha is null;
```

Result:
```
 percent_empty_commits
-----------------------
                 67.94
(1 row)
```

Some 68% of all commits from the 100 sampled repos are missing a commit body in their log.

# Missing commit messages per repo
Now, let's try how many commits are missing a commit body per repo sorted by percentage missing:

``` SQL
with commits_per_repo as(
  select count(distinct sha) as shas, gitrepo from log_filtered group by gitrepo
)
select l.gitrepo,
count(distinct l.sha) as empty_commit_bodies,
round((count(distinct l.sha)::decimal/(select shas from commits_per_repo c where c.gitrepo = l.gitrepo))*100, 2) as percent_empty_commits
from log_filtered l left join message m on l.sha = m.sha
where m.sha is null
group by l.gitrepo order by count(distinct l.sha) desc;
```

Result:
```
                    gitrepo                    | empty_commit_bodies | percent_empty_commits
-----------------------------------------------+---------------------+-----------------------
 /apache/incubator-echarts                     |                 965 |                 96.50
 /CyC2018/CS-Notes                             |                 959 |                 96.09
 /30-seconds/30-seconds-of-code                |                 936 |                 93.69
 /getify/You-Dont-Know-JS                      |                 923 |                 92.30
 /jwasham/coding-interview-university          |                 919 |                 91.90
 /Snailclimb/JavaGuide                         |                 916 |                 91.78
 /PanJiaChen/vue-element-admin                 |                 911 |                 91.10
 /mrdoob/three.js                              |                 906 |                 90.69
 /hakimel/reveal.js                            |                 904 |                 90.40
 /socketio/socket.io                           |                 896 |                 89.60
 /expressjs/express                            |                 883 |                 88.30
 /storybooks/storybook                         |                 879 |                 87.90
 /laravel/laravel                              |                 873 |                 87.30
 /ElemeFE/element                              |                 871 |                 87.10
 /atom/atom                                    |                 865 |                 86.59
 /vinta/awesome-python                         |                 859 |                 85.99
 /Microsoft/TypeScript                         |                 850 |                 85.00
 /Microsoft/vscode                             |                 848 |                 84.80
 /Semantic-Org/Semantic-UI                     |                 846 |                 84.60
 /iluwatar/java-design-patterns                |                 842 |                 84.37
 /vuejs/vue                                    |                 830 |                 83.00
 /moment/moment                                |                 826 |                 82.60
 /httpie/httpie                                |                 814 |                 81.40
 /webpack/webpack                              |                 814 |                 81.40
 /ant-design/ant-design                        |                 814 |                 81.40
 /electron/electron                            |                 809 |                 81.06
 /jquery/jquery                                |                 808 |                 80.80
 /pallets/flask                                |                 807 |                 80.70
 /mui-org/material-ui                          |                 804 |                 80.40
 /trekhleb/javascript-algorithms               |                 793 |                 88.60
 /nvbn/thefuck                                 |                 788 |                 78.80
 /goldbergyoni/nodebestpractices               |                 783 |                 78.46
 /twbs/bootstrap                               |                 783 |                 78.30
 /996icu/996.ICU                               |                 777 |                 77.86
 /reduxjs/redux                                |                 774 |                 77.40
 /airbnb/javascript                            |                 768 |                 76.80
 /shadowsocks/shadowsocks-windows              |                 766 |                 76.60
 /lodash/lodash                                |                 762 |                 76.20
 /avelino/awesome-go                           |                 760 |                 76.00
 /ReactiveX/RxJava                             |                 750 |                 75.00
 /chartjs/Chart.js                             |                 747 |                 74.70
 /kubernetes/kubernetes                        |                 743 |                 74.30
 /rust-lang/rust                               |                 738 |                 73.80
 /typicode/json-server                         |                 727 |                 96.04
 /adam-p/markdown-here                         |                 718 |                 97.03
 /d3/d3                                        |                 714 |                 71.40
 /denoland/deno                                |                 708 |                 70.80
 /Microsoft/PowerToys                          |                 705 |                 70.57
 /redis/redis                                  |                 704 |                 70.40
 /vuejs/awesome-vue                            |                 694 |                 69.40
 /ohmyzsh/ohmyzsh                              |                 678 |                 67.80
 /freeCodeCamp/freeCodeCamp                    |                 672 |                 67.61
 /netdata/netdata                              |                 670 |                 78.55
 /josephmisiti/awesome-machine-learning        |                 661 |                 78.69
 /h5bp/html5-boilerplate                       |                 659 |                 65.90
 /EbookFoundation/free-programming-books       |                 646 |                 64.60
 /sindresorhus/awesome                         |                 636 |                 79.20
 /jlevy/the-art-of-command-line                |                 635 |                 78.98
 /bitcoin/bitcoin                              |                 618 |                 61.80
 /facebook/create-react-app                    |                 612 |                 61.20
 /github/gitignore                             |                 607 |                 60.70
 /ossu/computer-science                        |                 586 |                 79.95
 /rails/rails                                  |                 572 |                 57.20
 /facebook/react                               |                 556 |                 55.60
 /ansible/ansible                              |                 553 |                 55.36
 /justjavac/free-programming-books-zh_CN       |                 544 |                 88.46
 /flutter/flutter                              |                 537 |                 53.75
 /zeit/next.js                                 |                 490 |                 49.00
 /gatsbyjs/gatsby                              |                 455 |                 45.59
 /tensorflow/models                            |                 435 |                 43.54
 /gohugoio/hugo                                |                 432 |                 43.20
 /angular/angular.js                           |                 424 |                 42.40
 /h5bp/Front-end-Developer-Interview-Questions |                 415 |                 75.32
 /tonsky/FiraCode                              |                 407 |                 91.46
 /GoThinkster/realworld                        |                 396 |                 92.52
 /kdn251/interviews                            |                 381 |                 95.97
 /TheAlgorithms/Python                         |                 372 |                 37.35
 /elastic/elasticsearch                        |                 363 |                 36.30
 /django/django                                |                 362 |                 36.27
 /apple/swift                                  |                 301 |                 30.10
 /ryanmcdermott/clean-code-javascript          |                 301 |                 80.48
 /danistefanovic/build-your-own-x              |                 299 |                 93.73
 /donnemartin/system-design-primer             |                 298 |                 96.75
 /kamranahmedse/developer-roadmap              |                 280 |                 92.41
 /tensorflow/tensorflow                        |                 264 |                 26.40
 /chrislgarry/Apollo-11                        |                 245 |                 56.32
 /angular/angular                              |                 233 |                 23.30
 /doocs/advanced-java                          |                 224 |                 49.23
 /Microsoft/terminal                           |                 211 |                 21.10
 /pytorch/pytorch                              |                 185 |                 18.50
 /spring-projects/spring-boot                  |                 180 |                 18.00
 /nodejs/node                                  |                 175 |                 17.50
 /moby/moby                                    |                 166 |                 16.60
 /MisterBooo/LeetCodeAnimation                 |                 159 |                 91.91
 /resume/resume.github.com                     |                 157 |                 88.20
 /facebook/react-native                        |                  90 |                  9.00
 /FortAwesome/Font-Awesome                     |                  67 |                 94.37
 /kelseyhightower/nocode                       |                   4 |                100.00
 /golang/go                                    |                   2 |                  0.20
 /torvalds/linux                               |                   1 |                  0.10
(100 rows)
```

A note on the nocode is that it's a repo on not writing code (to hammer home the important point on simplicity I guess?) and it contains 4 commits hence the 100% missing commit bodies.

# Wordyness of the repos
What's the distribution of the word count of commit messages? I'll limit the wordyness of the commits to a word count of 400 words to filter out outliers which are most probably  machine generated.

Over all repos:
```SQL
with message_count as (
  select count(*) from message
)
select width_bucket(length, 1, 400, 19) as bucket,
int4range(min(length), max(length), '[]') as range,
count(*) as freq,
round((count(*)::decimal/(select count from message_count))*100, 2) as percentage_of_commits
from message
where length <= 400
group by bucket
order by bucket desc;
```

Result:
```
 bucket |   range   | freq  | percentage_of_commits
--------+-----------+-------+-----------------------
     19 | [379,400) |    12 |                  0.04
     18 | [359,379) |    18 |                  0.06
     17 | [338,358) |    23 |                  0.08
     16 | [316,337) |    25 |                  0.09
     15 | [295,316) |    31 |                  0.11
     14 | [275,295) |    49 |                  0.17
     13 | [255,274) |    44 |                  0.16
     12 | [232,253) |    70 |                  0.25
     11 | [211,232) |    83 |                  0.29
     10 | [190,211) |    98 |                  0.35
      9 | [169,190) |   143 |                  0.50
      8 | [148,169) |   171 |                  0.60
      7 | [127,148) |   241 |                  0.85
      6 | [106,127) |   417 |                  1.47
      5 | [85,106)  |   657 |                  2.32
      4 | [64,85)   |  1158 |                  4.08
      3 | [43,64)   |  2248 |                  7.92
      2 | [22,43)   |  5278 |                 18.61
      1 | [1,22)    | 17434 |                 61.46
(19 rows)
```

61% commits consists of 1-22 words or less. Remember that this is from the actual 32% commits that had a body to begin with.

Just for the sake of it, let's check that lower bound in detail with a range of 1-22 divided into ten buckets:

```SQL
with message_count as (
  select count(*) from message_filtered
)
select width_bucket(length, 1, 22, 9) as bucket,
int4range(min(length), max(length), '[]') as range,
count(*) as freq,
round((count(*)::decimal/(select count from message_count))*100, 2) as percentage_of_commits
from message_filtered
where length <= 22
group by bucket
order by bucket desc;
```

Result:
```
 bucket |  range  | freq | percentage_of_commits
--------+---------+------+-----------------------
     10 | [22,23) |  348 |                  1.23
      9 | [20,22) |  864 |                  3.05
      8 | [18,20) |  946 |                  3.33
      7 | [15,18) | 1644 |                  5.80
      6 | [13,15) | 1154 |                  4.07
      5 | [11,13) | 1201 |                  4.23
      4 | [8,11)  | 2019 |                  7.12
      3 | [6,8)   | 1323 |                  4.66
      2 | [4,6)   | 2722 |                  9.60
      1 | [1,4)   | 5561 |                 19.60
(10 rows)
```

Let's widen the range to 1-50 and see that per repo:

```SQL
with message_count as (
  select count(*), gitrepo from message group by gitrepo
)
select width_bucket(m.length, 1, 50, 3) as bucket,
int4range(min(m.length), max(m.length), '[]') as range,
count(*) as freq,
m.gitrepo,
round((count(*)::decimal/(select count from message_count c where c.gitrepo = m.gitrepo))*100, 2) as percentage_of_commits
from message m
where length <= 50
group by m.gitrepo, bucket
order by m.gitrepo, bucket desc;
```

Result:
```
 bucket |  range  | freq |                    gitrepo                    | percentage_of_commits
--------+---------+------+-----------------------------------------------+-----------------------
      3 | [37,48) |    2 | /30-seconds/30-seconds-of-code                |                  3.13
      2 | [19,33) |    6 | /30-seconds/30-seconds-of-code                |                  9.38
      1 | [1,18)  |   54 | /30-seconds/30-seconds-of-code                |                 84.38
      3 | [36,40) |    4 | /996icu/996.ICU                               |                  1.80
      2 | [18,34) |   12 | /996icu/996.ICU                               |                  5.41
      1 | [1,18)  |  202 | /996icu/996.ICU                               |                 90.99
      3 | [38,43) |    5 | /adam-p/markdown-here                         |                 22.73
      2 | [20,34) |    5 | /adam-p/markdown-here                         |                 22.73
      1 | [4,16)  |    8 | /adam-p/markdown-here                         |                 36.36
      4 | [50,51) |    2 | /airbnb/javascript                            |                  0.86
      3 | [34,47) |   16 | /airbnb/javascript                            |                  6.90
      2 | [18,34) |   47 | /airbnb/javascript                            |                 20.26
      1 | [1,18)  |  139 | /airbnb/javascript                            |                 59.91
      4 | [50,51) |    1 | /angular/angular                              |                  0.13
      3 | [34,50) |   86 | /angular/angular                              |                 11.21
      2 | [18,34) |  116 | /angular/angular                              |                 15.12
      1 | [1,18)  |  417 | /angular/angular                              |                 54.37
      3 | [34,50) |   51 | /angular/angular.js                           |                  8.85
      2 | [18,34) |  120 | /angular/angular.js                           |                 20.83
      1 | [2,18)  |  305 | /angular/angular.js                           |                 52.95
      4 | [50,51) |    3 | /ansible/ansible                              |                  0.67
      3 | [34,50) |   47 | /ansible/ansible                              |                 10.54
      2 | [18,34) |  115 | /ansible/ansible                              |                 25.78
      1 | [1,18)  |  206 | /ansible/ansible                              |                 46.19
      3 | [35,49) |    9 | /ant-design/ant-design                        |                  4.84
      2 | [18,34) |   32 | /ant-design/ant-design                        |                 17.20
      1 | [1,18)  |  133 | /ant-design/ant-design                        |                 71.51
      2 | [18,26) |    2 | /apache/incubator-echarts                     |                  5.71
      1 | [1,18)  |   32 | /apache/incubator-echarts                     |                 91.43
      4 | [50,51) |    5 | /apple/swift                                  |                  0.72
      3 | [34,50) |   69 | /apple/swift                                  |                  9.87
      2 | [18,34) |  143 | /apple/swift                                  |                 20.46
      1 | [1,18)  |  351 | /apple/swift                                  |                 50.21
      3 | [34,47) |   12 | /atom/atom                                    |                  8.96
      2 | [18,33) |   29 | /atom/atom                                    |                 21.64
      1 | [2,18)  |   83 | /atom/atom                                    |                 61.94
      3 | [34,44) |    7 | /avelino/awesome-go                           |                  2.92
      2 | [18,34) |   33 | /avelino/awesome-go                           |                 13.75
      1 | [1,18)  |  171 | /avelino/awesome-go                           |                 71.25
      4 | [50,51) |    2 | /bitcoin/bitcoin                              |                  0.52
      3 | [34,50) |   54 | /bitcoin/bitcoin                              |                 14.14
      2 | [18,34) |   98 | /bitcoin/bitcoin                              |                 25.65
      1 | [1,18)  |  166 | /bitcoin/bitcoin                              |                 43.46
      4 | [50,51) |    3 | /chartjs/Chart.js                             |                  1.19
      3 | [34,48) |   30 | /chartjs/Chart.js                             |                 11.86
      2 | [18,32) |   63 | /chartjs/Chart.js                             |                 24.90
      1 | [1,18)  |  134 | /chartjs/Chart.js                             |                 52.96
      3 | [34,50) |   15 | /chrislgarry/Apollo-11                        |                  7.85
      2 | [18,34) |   40 | /chrislgarry/Apollo-11                        |                 20.94
      1 | [2,18)  |  117 | /chrislgarry/Apollo-11                        |                 61.26
      1 | [1,15)  |   39 | /CyC2018/CS-Notes                             |                100.00
      4 | [50,51) |    3 | /d3/d3                                        |                  1.05
      3 | [34,50) |   38 | /d3/d3                                        |                 13.29
      2 | [18,34) |   77 | /d3/d3                                        |                 26.92
      1 | [2,18)  |  111 | /d3/d3                                        |                 38.81
      2 | [20,25) |    4 | /danistefanovic/build-your-own-x              |                 20.00
      1 | [1,17)  |   16 | /danistefanovic/build-your-own-x              |                 80.00
      4 | [50,51) |    1 | /denoland/deno                                |                  0.34
      3 | [36,50) |   19 | /denoland/deno                                |                  6.51
      2 | [18,34) |   49 | /denoland/deno                                |                 16.78
      1 | [1,18)  |  197 | /denoland/deno                                |                 67.47
      3 | [34,50) |   23 | /django/django                                |                  3.61
      2 | [18,32) |   28 | /django/django                                |                  4.39
      1 | [2,18)  |  569 | /django/django                                |                 89.18
      2 | [23,26) |    2 | /donnemartin/system-design-primer             |                 20.00
      1 | [2,17)  |    8 | /donnemartin/system-design-primer             |                 80.00
      3 | [34,49) |    7 | /doocs/advanced-java                          |                  3.03
      2 | [18,31) |   17 | /doocs/advanced-java                          |                  7.36
      1 | [1,18)  |  197 | /doocs/advanced-java                          |                 85.28
      3 | [34,48) |   42 | /EbookFoundation/free-programming-books       |                 11.86
      2 | [18,34) |   96 | /EbookFoundation/free-programming-books       |                 27.12
      1 | [1,18)  |  189 | /EbookFoundation/free-programming-books       |                 53.39
      4 | [50,51) |   10 | /elastic/elasticsearch                        |                  1.57
      3 | [34,50) |   93 | /elastic/elasticsearch                        |                 14.60
      2 | [18,34) |  123 | /elastic/elasticsearch                        |                 19.31
      1 | [2,18)  |  247 | /elastic/elasticsearch                        |                 38.78
      4 | [50,51) |    2 | /electron/electron                            |                  1.05
      3 | [34,48) |   12 | /electron/electron                            |                  6.32
      2 | [18,34) |   37 | /electron/electron                            |                 19.47
      1 | [1,18)  |  117 | /electron/electron                            |                 61.58
      3 | [35,50) |    7 | /ElemeFE/element                              |                  5.43
      2 | [18,34) |   35 | /ElemeFE/element                              |                 27.13
      1 | [1,18)  |   81 | /ElemeFE/element                              |                 62.79
      4 | [50,51) |    1 | /expressjs/express                            |                  0.85
      3 | [37,46) |    3 | /expressjs/express                            |                  2.56
      2 | [18,34) |   12 | /expressjs/express                            |                 10.26
      1 | [1,18)  |   97 | /expressjs/express                            |                 82.91
      4 | [50,51) |    6 | /facebook/create-react-app                    |                  1.55
      3 | [34,49) |   34 | /facebook/create-react-app                    |                  8.76
      2 | [18,34) |  101 | /facebook/create-react-app                    |                 26.03
      1 | [1,18)  |  194 | /facebook/create-react-app                    |                 50.00
      4 | [50,51) |    3 | /facebook/react                               |                  0.68
      3 | [34,50) |   57 | /facebook/react                               |                 12.84
      2 | [18,34) |  101 | /facebook/react                               |                 22.75
      1 | [1,18)  |  177 | /facebook/react                               |                 39.86
      4 | [50,51) |    8 | /facebook/react-native                        |                  0.88
      3 | [34,50) |  117 | /facebook/react-native                        |                 12.86
      2 | [18,34) |  167 | /facebook/react-native                        |                 18.35
      1 | [2,18)  |  329 | /facebook/react-native                        |                 36.15
      4 | [50,51) |    2 | /flutter/flutter                              |                  0.43
      3 | [34,50) |   59 | /flutter/flutter                              |                 12.74
      2 | [18,34) |   95 | /flutter/flutter                              |                 20.52
      1 | [1,18)  |  195 | /flutter/flutter                              |                 42.12
      2 | [20,21) |    1 | /FortAwesome/Font-Awesome                     |                 25.00
      1 | [1,18)  |    3 | /FortAwesome/Font-Awesome                     |                 75.00
      3 | [34,50) |   26 | /freeCodeCamp/freeCodeCamp                    |                  8.00
      2 | [18,34) |   70 | /freeCodeCamp/freeCodeCamp                    |                 21.54
      1 | [1,18)  |  210 | /freeCodeCamp/freeCodeCamp                    |                 64.62
      4 | [50,51) |    6 | /gatsbyjs/gatsby                              |                  1.10
      3 | [34,50) |   51 | /gatsbyjs/gatsby                              |                  9.36
      2 | [18,34) |  122 | /gatsbyjs/gatsby                              |                 22.39
      1 | [1,18)  |  261 | /gatsbyjs/gatsby                              |                 47.89
      3 | [38,49) |    4 | /getify/You-Dont-Know-JS                      |                  5.19
      2 | [18,33) |   12 | /getify/You-Dont-Know-JS                      |                 15.58
      1 | [1,17)  |   55 | /getify/You-Dont-Know-JS                      |                 71.43
      4 | [50,51) |    3 | /github/gitignore                             |                  0.76
      3 | [34,50) |   44 | /github/gitignore                             |                 11.20
      2 | [18,34) |   97 | /github/gitignore                             |                 24.68
      1 | [1,18)  |  216 | /github/gitignore                             |                 54.96
      4 | [50,51) |    1 | /gohugoio/hugo                                |                  0.18
      3 | [34,50) |   38 | /gohugoio/hugo                                |                  6.69
      2 | [18,33) |   66 | /gohugoio/hugo                                |                 11.62
      1 | [2,18)  |  415 | /gohugoio/hugo                                |                 73.06
      4 | [50,51) |    3 | /golang/go                                    |                  0.30
      3 | [34,50) |  108 | /golang/go                                    |                 10.82
      2 | [18,34) |  176 | /golang/go                                    |                 17.64
      1 | [1,18)  |  441 | /golang/go                                    |                 44.19
      3 | [45,47) |    2 | /goldbergyoni/nodebestpractices               |                  0.93
      2 | [20,33) |    9 | /goldbergyoni/nodebestpractices               |                  4.19
      1 | [1,17)  |  200 | /goldbergyoni/nodebestpractices               |                 93.02
      4 | [50,51) |    1 | /GoThinkster/realworld                        |                  3.13
      2 | [19,28) |    6 | /GoThinkster/realworld                        |                 18.75
      1 | [1,18)  |   25 | /GoThinkster/realworld                        |                 78.13
      3 | [35,47) |    7 | /h5bp/Front-end-Developer-Interview-Questions |                  5.11
      2 | [18,34) |   43 | /h5bp/Front-end-Developer-Interview-Questions |                 31.39
      1 | [1,18)  |   83 | /h5bp/Front-end-Developer-Interview-Questions |                 60.58
      4 | [50,51) |    4 | /h5bp/html5-boilerplate                       |                  1.17
      3 | [34,50) |   30 | /h5bp/html5-boilerplate                       |                  8.80
      2 | [18,34) |   76 | /h5bp/html5-boilerplate                       |                 22.29
      1 | [1,18)  |  175 | /h5bp/html5-boilerplate                       |                 51.32
      3 | [37,50) |    8 | /hakimel/reveal.js                            |                  8.33
      2 | [19,34) |   24 | /hakimel/reveal.js                            |                 25.00
      1 | [2,18)  |   55 | /hakimel/reveal.js                            |                 57.29
      3 | [34,50) |   11 | /httpie/httpie                                |                  5.91
      2 | [18,34) |   25 | /httpie/httpie                                |                 13.44
      1 | [1,17)  |  141 | /httpie/httpie                                |                 75.81
      4 | [50,51) |    1 | /iluwatar/java-design-patterns                |                  0.64
      3 | [34,50) |   15 | /iluwatar/java-design-patterns                |                  9.62
      2 | [18,34) |   45 | /iluwatar/java-design-patterns                |                 28.85
      1 | [1,18)  |   79 | /iluwatar/java-design-patterns                |                 50.64
      3 | [34,47) |    5 | /jlevy/the-art-of-command-line                |                  2.96
      2 | [18,34) |   13 | /jlevy/the-art-of-command-line                |                  7.69
      1 | [1,18)  |  151 | /jlevy/the-art-of-command-line                |                 89.35
      3 | [34,50) |    9 | /josephmisiti/awesome-machine-learning        |                  5.03
      2 | [18,33) |   38 | /josephmisiti/awesome-machine-learning        |                 21.23
      1 | [1,18)  |  124 | /josephmisiti/awesome-machine-learning        |                 69.27
      4 | [50,51) |    1 | /jquery/jquery                                |                  0.52
      3 | [34,49) |   14 | /jquery/jquery                                |                  7.29
      2 | [18,33) |   35 | /jquery/jquery                                |                 18.23
      1 | [2,18)  |  121 | /jquery/jquery                                |                 63.02
      3 | [34,35) |    1 | /justjavac/free-programming-books-zh_CN       |                  1.41
      1 | [1,15)  |   69 | /justjavac/free-programming-books-zh_CN       |                 97.18
      3 | [36,42) |    3 | /jwasham/coding-interview-university          |                  3.70
      2 | [18,33) |    6 | /jwasham/coding-interview-university          |                  7.41
      1 | [1,18)  |   70 | /jwasham/coding-interview-university          |                 86.42
      3 | [35,47) |    6 | /kamranahmedse/developer-roadmap              |                 26.09
      2 | [19,31) |    4 | /kamranahmedse/developer-roadmap              |                 17.39
      1 | [3,18)  |   10 | /kamranahmedse/developer-roadmap              |                 43.48
      2 | [20,30) |    3 | /kdn251/interviews                            |                 18.75
      1 | [1,17)  |   13 | /kdn251/interviews                            |                 81.25
      3 | [34,49) |   27 | /kubernetes/kubernetes                        |                 10.51
      2 | [18,34) |   61 | /kubernetes/kubernetes                        |                 23.74
      1 | [1,18)  |  136 | /kubernetes/kubernetes                        |                 52.92
      3 | [35,47) |    4 | /laravel/laravel                              |                  3.15
      2 | [19,31) |   14 | /laravel/laravel                              |                 11.02
      1 | [1,18)  |  103 | /laravel/laravel                              |                 81.10
      3 | [49,50) |    2 | /lodash/lodash                                |                  0.84
      2 | [22,24) |    3 | /lodash/lodash                                |                  1.26
      1 | [2,17)  |  232 | /lodash/lodash                                |                 97.48
      4 | [50,51) |    1 | /Microsoft/PowerToys                          |                  0.34
      3 | [34,50) |   22 | /Microsoft/PowerToys                          |                  7.48
      2 | [18,34) |   68 | /Microsoft/PowerToys                          |                 23.13
      1 | [1,18)  |  150 | /Microsoft/PowerToys                          |                 51.02
      4 | [50,51) |    3 | /Microsoft/terminal                           |                  0.38
      3 | [34,50) |   70 | /Microsoft/terminal                           |                  8.87
      2 | [18,34) |  101 | /Microsoft/terminal                           |                 12.80
      1 | [2,18)  |  161 | /Microsoft/terminal                           |                 20.41
      3 | [34,50) |    9 | /Microsoft/TypeScript                         |                  6.00
      2 | [18,34) |   50 | /Microsoft/TypeScript                         |                 33.33
      1 | [2,18)  |   63 | /Microsoft/TypeScript                         |                 42.00
      3 | [35,48) |    8 | /Microsoft/vscode                             |                  5.26
      2 | [18,34) |   17 | /Microsoft/vscode                             |                 11.18
      1 | [1,17)  |  118 | /Microsoft/vscode                             |                 77.63
      1 | [1,3)   |   14 | /MisterBooo/LeetCodeAnimation                 |                100.00
      4 | [50,51) |    2 | /moby/moby                                    |                  0.24
      3 | [34,50) |   59 | /moby/moby                                    |                  7.07
      2 | [18,34) |  106 | /moby/moby                                    |                 12.71
      1 | [2,18)  |  594 | /moby/moby                                    |                 71.22
      4 | [50,51) |    1 | /moment/moment                                |                  0.57
      3 | [34,49) |   22 | /moment/moment                                |                 12.64
      2 | [18,34) |   28 | /moment/moment                                |                 16.09
      1 | [1,18)  |  110 | /moment/moment                                |                 63.22
      4 | [50,51) |    2 | /mrdoob/three.js                              |                  2.15
      3 | [34,50) |   13 | /mrdoob/three.js                              |                 13.98
      2 | [18,32) |   17 | /mrdoob/three.js                              |                 18.28
      1 | [1,18)  |   51 | /mrdoob/three.js                              |                 54.84
      4 | [50,51) |    3 | /mui-org/material-ui                          |                  1.53
      3 | [34,50) |   22 | /mui-org/material-ui                          |                 11.22
      2 | [18,32) |   51 | /mui-org/material-ui                          |                 26.02
      1 | [1,18)  |   99 | /mui-org/material-ui                          |                 50.51
      4 | [50,51) |    2 | /netdata/netdata                              |                  1.09
      3 | [35,50) |   32 | /netdata/netdata                              |                 17.49
      2 | [18,34) |   44 | /netdata/netdata                              |                 24.04
      1 | [2,18)  |   56 | /netdata/netdata                              |                 30.60
      4 | [50,51) |    6 | /nodejs/node                                  |                  0.73
      3 | [34,50) |  189 | /nodejs/node                                  |                 22.91
      2 | [18,34) |  277 | /nodejs/node                                  |                 33.58
      1 | [1,18)  |  193 | /nodejs/node                                  |                 23.39
      4 | [50,51) |    1 | /nvbn/thefuck                                 |                  0.47
      3 | [34,50) |   25 | /nvbn/thefuck                                 |                 11.79
      2 | [18,33) |   47 | /nvbn/thefuck                                 |                 22.17
      1 | [1,18)  |  112 | /nvbn/thefuck                                 |                 52.83
      4 | [50,51) |    1 | /ohmyzsh/ohmyzsh                              |                  0.31
      3 | [34,50) |   22 | /ohmyzsh/ohmyzsh                              |                  6.83
      2 | [18,34) |   57 | /ohmyzsh/ohmyzsh                              |                 17.70
      1 | [1,18)  |  202 | /ohmyzsh/ohmyzsh                              |                 62.73
      3 | [34,49) |   13 | /ossu/computer-science                        |                  8.84
      2 | [18,34) |   16 | /ossu/computer-science                        |                 10.88
      1 | [1,18)  |  109 | /ossu/computer-science                        |                 74.15
      4 | [50,51) |    1 | /pallets/flask                                |                  0.52
      3 | [36,50) |   14 | /pallets/flask                                |                  7.25
      2 | [18,34) |   40 | /pallets/flask                                |                 20.73
      1 | [1,18)  |  122 | /pallets/flask                                |                 63.21
      2 | [18,32) |    8 | /PanJiaChen/vue-element-admin                 |                  8.99
      1 | [1,18)  |   78 | /PanJiaChen/vue-element-admin                 |                 87.64
      4 | [50,51) |    1 | /pytorch/pytorch                              |                  0.12
      3 | [34,50) |  118 | /pytorch/pytorch                              |                 14.48
      2 | [18,34) |  286 | /pytorch/pytorch                              |                 35.09
      1 | [1,18)  |  189 | /pytorch/pytorch                              |                 23.19
      4 | [50,51) |    1 | /rails/rails                                  |                  0.23
      3 | [34,50) |   32 | /rails/rails                                  |                  7.48
      2 | [18,34) |   67 | /rails/rails                                  |                 15.65
      1 | [1,18)  |  272 | /rails/rails                                  |                 63.55
      4 | [50,51) |    1 | /ReactiveX/RxJava                             |                  0.40
      3 | [34,47) |   19 | /ReactiveX/RxJava                             |                  7.60
      2 | [18,34) |   65 | /ReactiveX/RxJava                             |                 26.00
      1 | [1,18)  |  145 | /ReactiveX/RxJava                             |                 58.00
      4 | [50,51) |    1 | /redis/redis                                  |                  0.34
      3 | [34,50) |   43 | /redis/redis                                  |                 14.53
      2 | [18,34) |   75 | /redis/redis                                  |                 25.34
      1 | [2,18)  |   86 | /redis/redis                                  |                 29.05
      3 | [34,50) |   28 | /reduxjs/redux                                |                 12.39
      2 | [18,33) |   60 | /reduxjs/redux                                |                 26.55
      1 | [1,18)  |  111 | /reduxjs/redux                                |                 49.12
      3 | [35,47) |    2 | /resume/resume.github.com                     |                  9.52
      2 | [27,30) |    4 | /resume/resume.github.com                     |                 19.05
      1 | [1,18)  |   14 | /resume/resume.github.com                     |                 66.67
      4 | [50,51) |    1 | /rust-lang/rust                               |                  0.38
      3 | [34,50) |   25 | /rust-lang/rust                               |                  9.54
      2 | [18,34) |   55 | /rust-lang/rust                               |                 20.99
      1 | [1,18)  |  138 | /rust-lang/rust                               |                 52.67
      3 | [41,50) |    5 | /ryanmcdermott/clean-code-javascript          |                  6.85
      2 | [18,33) |   18 | /ryanmcdermott/clean-code-javascript          |                 24.66
      1 | [2,18)  |   42 | /ryanmcdermott/clean-code-javascript          |                 57.53
      3 | [46,47) |    2 | /Semantic-Org/Semantic-UI                     |                  1.30
      2 | [19,30) |    3 | /Semantic-Org/Semantic-UI                     |                  1.95
      1 | [2,17)  |  147 | /Semantic-Org/Semantic-UI                     |                 95.45
      3 | [34,49) |    7 | /shadowsocks/shadowsocks-windows              |                  2.99
      2 | [19,33) |   33 | /shadowsocks/shadowsocks-windows              |                 14.10
      1 | [1,18)  |  184 | /shadowsocks/shadowsocks-windows              |                 78.63
      3 | [36,43) |    7 | /sindresorhus/awesome                         |                  4.19
      2 | [18,34) |   21 | /sindresorhus/awesome                         |                 12.57
      1 | [1,18)  |  133 | /sindresorhus/awesome                         |                 79.64
      2 | [18,19) |    1 | /Snailclimb/JavaGuide                         |                  1.22
      1 | [1,14)  |   81 | /Snailclimb/JavaGuide                         |                 98.78
      3 | [37,49) |    9 | /socketio/socket.io                           |                  8.65
      2 | [18,34) |   22 | /socketio/socket.io                           |                 21.15
      1 | [1,18)  |   57 | /socketio/socket.io                           |                 54.81
      4 | [50,51) |    1 | /spring-projects/spring-boot                  |                  0.12
      3 | [34,50) |   53 | /spring-projects/spring-boot                  |                  6.46
      2 | [18,34) |   85 | /spring-projects/spring-boot                  |                 10.37
      1 | [1,18)  |  609 | /spring-projects/spring-boot                  |                 74.27
      3 | [36,50) |    9 | /storybooks/storybook                         |                  7.44
      2 | [19,33) |   13 | /storybooks/storybook                         |                 10.74
      1 | [1,18)  |   94 | /storybooks/storybook                         |                 77.69
      4 | [50,51) |    1 | /tensorflow/models                            |                  0.18
      3 | [34,50) |   21 | /tensorflow/models                            |                  3.72
      2 | [18,34) |   51 | /tensorflow/models                            |                  9.04
      1 | [2,18)  |  456 | /tensorflow/models                            |                 80.85
      4 | [50,51) |    3 | /tensorflow/tensorflow                        |                  0.41
      3 | [34,50) |   48 | /tensorflow/tensorflow                        |                  6.52
      2 | [18,34) |   80 | /tensorflow/tensorflow                        |                 10.87
      1 | [2,18)  |  525 | /tensorflow/tensorflow                        |                 71.33
      4 | [50,51) |    2 | /TheAlgorithms/Python                         |                  0.32
      3 | [34,50) |   91 | /TheAlgorithms/Python                         |                 14.54
      2 | [18,34) |  181 | /TheAlgorithms/Python                         |                 28.91
      1 | [1,18)  |  217 | /TheAlgorithms/Python                         |                 34.66
      3 | [42,43) |    1 | /tonsky/FiraCode                              |                  2.63
      2 | [20,34) |    9 | /tonsky/FiraCode                              |                 23.68
      1 | [1,18)  |   24 | /tonsky/FiraCode                              |                 63.16
      4 | [50,51) |   14 | /torvalds/linux                               |                  1.40
      3 | [34,50) |  198 | /torvalds/linux                               |                 19.82
      2 | [18,34) |  249 | /torvalds/linux                               |                 24.92
      1 | [4,18)  |  127 | /torvalds/linux                               |                 12.71
      4 | [50,51) |    1 | /trekhleb/javascript-algorithms               |                  0.97
      3 | [37,50) |    5 | /trekhleb/javascript-algorithms               |                  4.85
      2 | [18,34) |   26 | /trekhleb/javascript-algorithms               |                 25.24
      1 | [2,18)  |   59 | /trekhleb/javascript-algorithms               |                 57.28
      3 | [34,50) |   14 | /twbs/bootstrap                               |                  6.45
      2 | [18,34) |   46 | /twbs/bootstrap                               |                 21.20
      1 | [1,18)  |  140 | /twbs/bootstrap                               |                 64.52
      3 | [39,43) |    2 | /typicode/json-server                         |                  6.67
      2 | [19,31) |    9 | /typicode/json-server                         |                 30.00
      1 | [2,15)  |   15 | /typicode/json-server                         |                 50.00
      4 | [50,51) |    1 | /vinta/awesome-python                         |                  0.71
      3 | [34,45) |    5 | /vinta/awesome-python                         |                  3.57
      2 | [18,34) |   17 | /vinta/awesome-python                         |                 12.14
      1 | [1,18)  |  110 | /vinta/awesome-python                         |                 78.57
      3 | [34,45) |   10 | /vuejs/awesome-vue                            |                  3.27
      2 | [18,34) |   36 | /vuejs/awesome-vue                            |                 11.76
      1 | [1,18)  |  254 | /vuejs/awesome-vue                            |                 83.01
      4 | [50,51) |    1 | /vuejs/vue                                    |                  0.59
      3 | [37,47) |   10 | /vuejs/vue                                    |                  5.88
      2 | [18,34) |   29 | /vuejs/vue                                    |                 17.06
      1 | [2,18)  |  119 | /vuejs/vue                                    |                 70.00
      3 | [39,41) |    5 | /webpack/webpack                              |                  2.69
      2 | [18,34) |   23 | /webpack/webpack                              |                 12.37
      1 | [1,18)  |  154 | /webpack/webpack                              |                 82.80
      4 | [50,51) |    2 | /zeit/next.js                                 |                  0.39
      3 | [34,49) |   58 | /zeit/next.js                                 |                 11.37
      2 | [18,34) |  140 | /zeit/next.js                                 |                 27.45
      1 | [1,18)  |  221 | /zeit/next.js                                 |                 43.33
(332 rows)
```

# Correlation of commit length and commit size
The last thing that would be interesting to see is how the length of the commits and the length of the commit-message correlates:

```SQL
with amount_change as  (
  select sum(added) + sum(removed) as amount_change, sha from log_filtered group by sha
)
select corr(a.amount_change, m.length)
from amount_change a inner join message_filtered m on a.sha = m.sha;
```

Result:
```
         corr
----------------------
 0.014841633941807921
(1 row)
```

And grouped per repo and ordered by correlation. The commit count for each repo has been included so outliers can be tempered by the count of commits:

```SQL
with amount_change as  (
  select sum(added) + sum(removed) as amount_change, sha from log group by sha
), message_count as (
  select count(*), gitrepo from message group by gitrepo
)
select corr(a.amount_change, m.length),
m.gitrepo,
(select count from message_count c where c.gitrepo = m.gitrepo) as commit_counts
from message m inner join amount_change a on m.sha = a.sha
group by m.gitrepo
order by corr(a.amount_change, m.length) desc;
```

Result:
```
          corr          |                    gitrepo                    | commit_counts
------------------------+-----------------------------------------------+---------------
     0.9574372524842768 | /PanJiaChen/vue-element-admin                 |            89
     0.8450233814735529 | /justjavac/free-programming-books-zh_CN       |            71
     0.7759785983536915 | /tonsky/FiraCode                              |            38
      0.756532003507315 | /ElemeFE/element                              |           129
     0.7015779401292735 | /Microsoft/TypeScript                         |           150
     0.4912870649927109 | /vuejs/vue                                    |           170
     0.4902984155388296 | /gohugoio/hugo                                |           568
    0.46763942812031256 | /MisterBooo/LeetCodeAnimation                 |            14
     0.4313365492178031 | /iluwatar/java-design-patterns                |           156
    0.41033981075527015 | /EbookFoundation/free-programming-books       |           354
     0.3838025687116204 | /zeit/next.js                                 |           510
    0.37277533242774435 | /Microsoft/vscode                             |           152
     0.3472356890921223 | /expressjs/express                            |           117
    0.31645149601538164 | /moment/moment                                |           174
     0.3097066501462905 | /facebook/create-react-app                    |           388
     0.3037516904914518 | /ant-design/ant-design                        |           186
     0.3029444219392455 | /kamranahmedse/developer-roadmap              |            23
     0.2809436669282958 | /mrdoob/three.js                              |            93
     0.2588541532044764 | /mui-org/material-ui                          |           196
    0.24141531553866796 | /Microsoft/PowerToys                          |           294
    0.23811275483635225 | /goldbergyoni/nodebestpractices               |           215
    0.21103043659532458 | /apple/swift                                  |           699
    0.20564957910863005 | /chartjs/Chart.js                             |           253
      0.205102733871489 | /TheAlgorithms/Python                         |           626
    0.19242379462658368 | /h5bp/Front-end-Developer-Interview-Questions |           137
     0.1728446928573801 | /freeCodeCamp/freeCodeCamp                    |           325
    0.16483533145425533 | /electron/electron                            |           190
    0.16352545807773672 | /elastic/elasticsearch                        |           637
     0.1534123165159232 | /doocs/advanced-java                          |           231
    0.15231959965167385 | /rust-lang/rust                               |           262
    0.13415000703306085 | /gatsbyjs/gatsby                              |           545
    0.13229193489983715 | /denoland/deno                                |           292
    0.13163372898264047 | /ansible/ansible                              |           446
    0.12476670835088555 | /nvbn/thefuck                                 |           212
     0.1222573244616053 | /chrislgarry/Apollo-11                        |           191
    0.12023623802952621 | /GoThinkster/realworld                        |            32
    0.11978087842784842 | /resume/resume.github.com                     |            21
    0.11727319781426875 | /996icu/996.ICU                               |           222
    0.11083055533930299 | /angular/angular                              |           767
    0.11062591654251819 | /d3/d3                                        |           286
    0.10638468885130116 | /angular/angular.js                           |           576
    0.10281431622180907 | /trekhleb/javascript-algorithms               |           103
    0.10026268961781298 | /netdata/netdata                              |           183
    0.09973009901272063 | /airbnb/javascript                            |           232
     0.0993723390907912 | /vuejs/awesome-vue                            |           306
    0.09221829608761511 | /getify/You-Dont-Know-JS                      |            77
    0.08521249300727192 | /shadowsocks/shadowsocks-windows              |           234
    0.06272758747798197 | /webpack/webpack                              |           186
   0.058003677940542175 | /httpie/httpie                                |           186
    0.05715677039848148 | /typicode/json-server                         |            30
   0.054954338489510514 | /github/gitignore                             |           393
   0.050513552852880866 | /avelino/awesome-go                           |           240
     0.0430170909297431 | /golang/go                                    |           998
    0.03688327875163765 | /redis/redis                                  |           296
    0.03500791736218847 | /adam-p/markdown-here                         |            22
     0.0333375486971471 | /pytorch/pytorch                              |           815
    0.03273879914012796 | /hakimel/reveal.js                            |            96
      0.029720349711297 | /moby/moby                                    |           834
    0.02829624970622859 | /torvalds/linux                               |           999
    0.02782854013880664 | /reduxjs/redux                                |           226
   0.025715207177517463 | /facebook/react-native                        |           910
   0.022812538135825757 | /spring-projects/spring-boot                  |           820
   0.021199143329940974 | /rails/rails                                  |           428
   0.020188872368983216 | /storybooks/storybook                         |           121
   0.016487060017009354 | /CyC2018/CS-Notes                             |            39
    0.01634671511821379 | /atom/atom                                    |           134
    0.01589736069225779 | /pallets/flask                                |           193
   0.009940200687530768 | /30-seconds/30-seconds-of-code                |            64
   0.008712790097149365 | /tensorflow/models                            |           564
   0.006283368154793595 | /donnemartin/system-design-primer             |            10
  0.0019569047084698407 | /tensorflow/tensorflow                        |           736
  0.0012476075055534008 | /facebook/react                               |           444
  0.0007487970683038227 | /flutter/flutter                              |           463
 -0.0011489564740620624 | /apache/incubator-echarts                     |            35
 -0.0019503722290534922 | /nodejs/node                                  |           825
 -0.0031084658022554465 | /twbs/bootstrap                               |           217
  -0.004560986120737108 | /ReactiveX/RxJava                             |           250
  -0.004605559884709319 | /django/django                                |           638
  -0.007265047330518219 | /Snailclimb/JavaGuide                         |            82
   -0.00800299482164321 | /danistefanovic/build-your-own-x              |            20
  -0.008036463239404108 | /Microsoft/terminal                           |           789
   -0.01122728992208422 | /kubernetes/kubernetes                        |           257
  -0.013974234042948626 | /bitcoin/bitcoin                              |           382
  -0.021014420931780575 | /lodash/lodash                                |           238
  -0.025093455934156617 | /Semantic-Org/Semantic-UI                     |           154
   -0.02534308801613047 | /jlevy/the-art-of-command-line                |           169
  -0.032428485313003325 | /h5bp/html5-boilerplate                       |           341
   -0.03993281481851439 | /sindresorhus/awesome                         |           167
  -0.044468805944314675 | /jquery/jquery                                |           192
   -0.04984839389339831 | /laravel/laravel                              |           127
  -0.053457907116712465 | /socketio/socket.io                           |           104
   -0.05977028399479585 | /vinta/awesome-python                         |           140
  -0.060609675241894166 | /ohmyzsh/ohmyzsh                              |           322
   -0.06806512151244953 | /ryanmcdermott/clean-code-javascript          |            73
   -0.07073943084163321 | /josephmisiti/awesome-machine-learning        |           179
   -0.10819625275966732 | /ossu/computer-science                        |           147
   -0.13163755367727653 | /jwasham/coding-interview-university          |            81
    -0.3666075718336376 | /kdn251/interviews                            |            16
   -0.37113480951260275 | /FortAwesome/Font-Awesome                     |             4
(99 rows)
```

# Dogfooding
How does my repo measure up to my own preachy standards? I'll cherry-pick my [sider](https://github.com/jonaslu/sider) repo as it has been increasingly written in the atomic literate commit style in mind. Same SQL-statements as above has been used so I'll omit these this second time around.

# Atomicity of my commits
Running the SQL to find out atomicity without excluding outliers:
```
 bucket |    range    | freq | percent_of_commits
--------+-------------+------+--------------------
     20 | [4003,4004) |    1 |             0.4367
     14 | [2744,2745) |    1 |             0.4367
      9 | [1775,1776) |    1 |             0.4367
      4 | [819,820)   |    1 |             0.4367
      3 | [433,525)   |    3 |             1.3100
      2 | [213,367)   |    8 |             3.4934
      1 | [1,209)     |  213 |            93.0131
      0 | [0,1)       |    1 |             0.4367
(8 rows)
```

Limiting the range of changes from 1-2000 to exclude outliers to get a clearer picture:
```
 bucket |    range    | freq | percent_of_commits
--------+-------------+------+--------------------
     88 | [1775,1776) |    1 |             0.4405
     41 | [819,820)   |    1 |             0.4405
     26 | [524,525)   |    2 |             0.8811
     22 | [433,434)   |    1 |             0.4405
     19 | [366,367)   |    1 |             0.4405
     18 | [359,360)   |    1 |             0.4405
     16 | [309,310)   |    1 |             0.4405
     15 | [293,294)   |    1 |             0.4405
     12 | [231,241)   |    2 |             0.8811
     11 | [205,216)   |    4 |             1.7621
     10 | [185,203)   |    3 |             1.3216
      9 | [164,168)   |    2 |             0.8811
      8 | [143,151)   |    4 |             1.7621
      7 | [126,143)   |    4 |             1.7621
      5 | [82,102)    |   12 |             5.2863
      4 | [62,82)     |   18 |             7.9295
      3 | [42,62)     |   34 |            14.9780
      2 | [22,42)     |   45 |            19.8238
      1 | [1,22)      |   89 |            39.2070
      0 | [0,1)       |    1 |             0.4405
(20 rows)
```

# Missing commit messages
 Empty commit-messages:
```
 percent_empty_commits
-----------------------
                  1.75
```

One of these are from a merge from a second commiter, the rest are yours truly.

# Wordyness of the repo
Word count of commit messages in 20 buckets.
```
 bucket |   range   | freq | percentage_of_commits
--------+-----------+------+-----------------------
     13 | [259,260) |    1 |                  0.44
     11 | [212,230) |    2 |                  0.89
     10 | [191,196) |    2 |                  0.89
      9 | [180,190) |    2 |                  0.89
      8 | [148,169) |    9 |                  4.00
      7 | [132,145) |    5 |                  2.22
      6 | [106,123) |    9 |                  4.00
      5 | [86,106)  |   15 |                  6.67
      4 | [64,85)   |   30 |                 13.33
      3 | [43,64)   |   47 |                 20.89
      2 | [22,43)   |   43 |                 19.11
      1 | [4,22)    |   60 |                 26.67
(12 rows)
```

# Correlation of commit length and commit size
And the correlation:
```
        corr         |    gitrepo
---------------------+---------------
 0.10442932835874538 | jonaslu/sider
(1 row)
```

# Discussion
The good news is that having  small commits seems to be commonplace. If they are atomic is hard to say but making atomic commits when the majority are large changes would be very hard and go against the atomic commit notion. Some atomic commits can of course be large as long as they change one thing in many places. But then that would probably be a sign that your code is not so de-coupled as possible and could use a good refactoring. Small commits are a prerequisite to atomic commits and we can't actually judge the atomicity without looking at the individual commits. So as far as the statistcs are concerned the situation is good.

Now to the bad news. The first thing to notice is how common it is for commits to not have a commit-body. A whopping 68% of all sampled commits missed a commit-body.

The shining examples were go and linux. Then it's a pretty big climb up to the middle tier.

As for wordyness, it's also low. 62% were 1-22 words and of these were 19% 1-4 words.

I've used word count as a proxy for how literate and helpful in understanding the change and commit. Is it a good proxy? It's a good negative proxy. If the word count in your commit body is low it's very hard to be literate. Only as the count goes up is there a chance of litteracy. As we've seen the majority are small (for example 20% of all commits were 1-4 words), so we can safely draw the conclusion that the literate style of commits is not that common.

# Dogfooding discussion
Now, what about my own repo? First I'll admit after scanning the log that I too have skimped on the litterate part at times. I said at the outset that it was a progressively literate commit log. I have not taken it to the extreme - but will do so even more on newer repos.

But let's start with the atomicity. Surprisingly my commits are actually larger than the commit size of all the 100-repos. In the 100-list 62% of the commits were in the range 1-22, but in mine only 39% are. This could be a sign of the lack of maturity in my repo. There has been little to no bug-fixing which I assume would bring the number up in the lower-range. Nevertheless most commits fall in the 4-85 size range indicating that lower numbers are good for atomicity.

What about the litterate part? 4 messages without commit-body is shame on me. One wasn't mine, but I could still argue that the author would redo the PR before merging.

Here we can see that the distribution of the number of words are higher than the other repos. So wordyness is a prerequisite to literacy, but of course not a garantee since we are using numbers instead of the real thing.

I was surprised at the low correlation between commit size and wordyness of the commit. I was expecting a higher count. Especially since there have been few bug-fixes (which tend to have long messages but change a few lines after much head-scratching).

# Closing words
So the state of commits in the wild seems to be the following: fortunately atomicity is good and without atomicity literacy wouldn't help anyway. What could be improved is the literacy of commits. A whopping 68% of the commits contained no message body whatsoever which tells me we're wasting precious meta-data which could improve understanding.
