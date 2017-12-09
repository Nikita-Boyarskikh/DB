ALTER TABLE forum_users DROP CONSTRAINT IF EXISTS fk__forum_users_forumID__forums_slug;
ALTER TABLE threads DROP CONSTRAINT IF EXISTS fk__threads_forumID__forums_slug;
ALTER TABLE voices DROP CONSTRAINT IF EXISTS fk__voices_userID__users_nickname;

DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS forums;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS voices;
DROP TABLE IF EXISTS forum_users;

DROP INDEX IF EXISTS idx__users_email_hash;

DROP INDEX IF EXISTS idx__forums_slug;

DROP INDEX IF EXISTS idx__threads_slug;
DROP INDEX IF EXISTS idx__threads_created;
DROP INDEX IF EXISTS idx__threads_forumID;

DROP INDEX IF EXISTS idx__posts_forumID;
DROP INDEX IF EXISTS idx__posts_authorID;
DROP INDEX IF EXISTS idx__posts_threadID;
DROP INDEX IF EXISTS idx__posts_parentID;
DROP INDEX IF EXISTS idx__posts_parents_gin;
DROP INDEX IF EXISTS idx__posts_created;
DROP INDEX IF EXISTS idx__posts_ID_threadID_parentID;
DROP INDEX IF EXISTS idx__posts_ID_threadID;
