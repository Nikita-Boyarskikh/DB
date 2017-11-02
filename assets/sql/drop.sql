DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS threads;
DROP TABLE IF EXISTS forums;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS voices;

DROP INDEX IF EXISTS idx__users_email;

DROP INDEX IF EXISTS idx__forums_slug;

DROP INDEX IF EXISTS idx__threads_slug;
DROP INDEX IF EXISTS idx__threads_created;
DROP INDEX IF EXISTS idx__threads_forumID;

DROP INDEX IF EXISTS idx__posts_forumID;
DROP INDEX IF EXISTS idx__posts_authorID;
DROP INDEX IF EXISTS idx__posts_threadID;
DROP INDEX IF EXISTS idx__posts_parentID;
DROP INDEX IF EXISTS idx__posts_created;
DROP INDEX IF EXISTS idx__posts_parents;