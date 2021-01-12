CREATE TABLE IF NOT EXISTS log (
  sha TEXT,
  date TIMESTAMPTZ,
  author TEXT,
  added INT,
  removed INT,
  filename TEXT,
  gitrepo TEXT,

  PRIMARY KEY (sha, filename)
);

CREATE TABLE IF NOT EXISTS message  (
  sha TEXT PRIMARY KEY,
  message TEXT,
  length INT,
  gitrepo TEXT
);
