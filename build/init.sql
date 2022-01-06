CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users(
    email citext UNIQUE NOT NULL,
    fullname varchar NOT NULL,
    nickname citext COLLATE ucs_basic UNIQUE PRIMARY KEY,
    about text NOT NULL DEFAULT ''
);
--оставить оба.
CREATE INDEX IF NOT EXISTS  users_email ON users(email); --ускорили вставку постов
CREATE INDEX IF NOT EXISTS  users_email_nickname ON users(email, nickname);  --ускорили вставку постов

CREATE UNLOGGED TABLE forums (
    title varchar NOT NULL,
    author citext references users(nickname),
    slug citext PRIMARY KEY,
    posts int DEFAULT 0,
    threads int DEFAULT 0
);
CREATE INDEX IF NOT EXISTS forums_users ON forums(author); --замедлило вставку постов, ускорило всё остальное

CREATE UNLOGGED TABLE forum_users (
    nickname citext references users(nickname),
    forum citext references forums(slug),
    CONSTRAINT fk UNIQUE(nickname, forum)
);
CREATE INDEX IF NOT EXISTS forum_users_nickname ON forum_users(nickname);
CREATE INDEX IF NOT EXISTS forum_users_forum ON forum_users(forum);
CREATE INDEX IF NOT EXISTS forum_users_full ON forum_users(nickname,forum);

CREATE UNLOGGED TABLE threads (
    id serial PRIMARY KEY,
    author citext references users(nickname),
    message citext NOT NULL,
    title citext NOT NULL,
    created_at timestamp with time zone,
    forum citext references forums(slug),
    slug citext,
    votes int
);
CREATE INDEX IF NOT EXISTS threads_forum ON threads(forum); --не убирать
CREATE INDEX IF NOT EXISTS threads_created_forum_index ON threads(forum, created_at);
CREATE INDEX  IF NOT EXISTS threads_cluster_thread ON threads(id, forum); --ускоряет
CREATE INDEX IF NOT EXISTS threads_search_full ON  threads(slug, id, forum);

CREATE UNLOGGED TABLE posts (
    id serial PRIMARY KEY ,
    author citext references users(nickname),
    post citext NOT NULL,
    created_at timestamp with time zone,
    forum citext references forums(slug),
    isEdited bool,
    parent int,
    thread int references threads(id),
    path  INTEGER[]
);

CREATE INDEX IF NOT EXISTS posts_thread ON posts(thread); --не убирать
CREATE INDEX IF NOT EXISTS posts_search_desc ON posts(thread, path DESC);
CREATE INDEX IF NOT EXISTS posts_search_asc ON posts(thread, path ASC);
CREATE INDEX IF NOT EXISTS posts_parent_thread_index ON posts(parent, thread);
CREATE INDEX IF NOT EXISTS posts_threads_ida ON posts(thread, id ASC);

CREATE INDEX IF NOT EXISTS  parent_tree_index
    ON posts ((path[1]), path DESC, id);
CREATE INDEX IF NOT EXISTS  parent_tree_index2
    ON posts ((path[1]), path ASC, id);
CREATE INDEX IF NOT EXISTS  parent_tree_index3
    ON posts (id, (path[1]) ASC);
CREATE INDEX IF NOT EXISTS  parent_tree_index4
    ON posts (id, (path[1]) DESC);

CREATE UNLOGGED TABLE votes (
    author citext references users(nickname),
    vote int,
    thread int references threads(id),
    CONSTRAINT checks UNIQUE(author, thread)
);
CREATE INDEX IF NOT EXISTS votes_full ON votes(author, vote, thread);

--dark magic
CREATE OR REPLACE FUNCTION update_path() RETURNS TRIGGER AS
$update_path$
DECLARE
parent_path  INTEGER[];
    parent_thread int;
BEGIN
SELECT path FROM posts WHERE id = new.parent INTO parent_path;
NEW.path := parent_path || new.id;
RETURN new;
END
$update_path$ LANGUAGE plpgsql;


CREATE TRIGGER path_update_trigger
    BEFORE INSERT
    ON posts
    FOR EACH ROW
    EXECUTE PROCEDURE update_path();