CREATE TABLE IF NOT EXISTS users (
  nickname CITEXT,
  fullname TEXT,
  email    CITEXT,
  about    TEXT
);

CREATE TABLE IF NOT EXISTS forums (
  ID      SERIAL,
  posts   INT8 DEFAULT 0,
  slug    CITEXT,
  threads INT4 DEFAULT 0,
  title   TEXT,
  userID  CITEXT
);

CREATE TABLE IF NOT EXISTS threads (
  ID       SERIAL4,
  authorID CITEXT,
  created  TIMESTAMPTZ(3) DEFAULT now(),
  forumID  CITEXT,
  message  TEXT,
  title    TEXT,
  slug     CITEXT,
  votes    INT4           DEFAULT 0
);

CREATE TABLE IF NOT EXISTS posts (
  ID       SERIAL8,
  authorID CITEXT,
  created  TIMESTAMP(3) DEFAULT now(),
  forumID  CITEXT,
  isEdited BOOLEAN      DEFAULT FALSE,
  message  TEXT,
  parentID INT8         DEFAULT 0,
  threadID INT4,
  parents  INT8 []
);

CREATE TABLE IF NOT EXISTS voices (
  threadID INT4,
  userID   CITEXT,
  voice    INT2
);

CREATE INDEX IF NOT EXISTS idx__users_nickname
  ON users (nickname);
CREATE INDEX IF NOT EXISTS idx__users_email
  ON users (email);

CREATE INDEX IF NOT EXISTS idx__forums_ID
  ON forums (ID);
CREATE INDEX IF NOT EXISTS idx__forums_slug
  ON forums (slug);

CREATE INDEX IF NOT EXISTS idx__threads_ID
  ON threads (ID);
CREATE INDEX IF NOT EXISTS idx__threads_slug
  ON threads (slug)
  WHERE slug IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx__voices_threadID_userID
  ON voices (threadID, userID);