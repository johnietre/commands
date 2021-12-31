CREATE TABLE IF NOT EXISTS words (
  word TEXT PRIMARY KEY,
  mastered INTEGER
);

CREATE TABLE IF NOT EXISTS definitions (
  word TEXT,
  definition TEXT,
  CONSTRAINT fk_word FOREIGN KEY (word) REFERENCES words (word) ON DELETE CASCADE
);
