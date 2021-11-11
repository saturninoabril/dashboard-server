CREATE TABLE IF NOT EXISTS {{.prefix}}system (
    key     VARCHAR(64) PRIMARY KEY,
    value   VARCHAR(1024) NULL
);

CREATE TABLE IF NOT EXISTS {{.prefix}}user (
    id          CHAR(26) PRIMARY KEY,
    create_at   BIGINT NOT NULL,
    update_at   BIGINT NOT NULL,
    email       VARCHAR(64) NOT NULL UNIQUE,
    email_verified  BOOLEAN NOT NULL,
    password    VARCHAR(128) NOT NULL,
    first_name  VARCHAR(64) NOT NULL,
    last_name   VARCHAR(64) NOT NULL,
    state       TEXT DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS {{.prefix}}session (
    id          CHAR(26) PRIMARY KEY,
    token       CHAR(26) NOT NULL,
    create_at   BIGINT NOT NULL,
    expires_at  BIGINT NOT NULL,
    user_id     VARCHAR(64) NOT NULL,
    csrf_token  CHAR(26) NOT NULL
);

CREATE TABLE IF NOT EXISTS {{.prefix}}token (
    token       VARCHAR(64) PRIMARY KEY,
    create_at   BIGINT DEFAULT NULL,
    type        VARCHAR(64) DEFAULT NULL,
    extra       VARCHAR(256) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS {{.prefix}}role (
    id          CHAR(26) PRIMARY KEY,
    name	    TEXT NOT NULL UNIQUE,
    create_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
    update_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000)
);

CREATE TABLE IF NOT EXISTS {{.prefix}}user_role (
    user_id     CHAR(26) NOT NULL,
    role_id		CHAR(26) NOT NULL,
    create_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
    update_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),

    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS {{.prefix}}oauthstate (
    id          CHAR(26) PRIMARY KEY,
    token       CHAR(26) NOT NULL,
    create_at   BIGINT NOT NULL,
    expires_at  BIGINT NOT NULL
);
