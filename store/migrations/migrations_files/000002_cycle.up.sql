CREATE TABLE IF NOT EXISTS {{.prefix}}cycles (
    id          CHAR(26) PRIMARY KEY,
    repo	    VARCHAR(64) NOT NULL,
    branch	    VARCHAR(64) NOT NULL,
    build	    VARCHAR(64) NOT NULL,
    state	    VARCHAR(64) DEFAULT NULL,
    specs_registered    SMALLINT NOT NULL DEFAULT 0,
    specs_done  SMALLINT NOT NULL DEFAULT 0,
    duration    BIGINT NOT NULL DEFAULT 0,
    pass        SMALLINT NOT NULL DEFAULT 0,
    fail        SMALLINT NOT NULL DEFAULT 0,
    pending     SMALLINT NOT NULL DEFAULT 0,
    skipped     SMALLINT NOT NULL DEFAULT 0,
    start_at    BIGINT DEFAULT NULL,
    end_at      BIGINT DEFAULT NULL,
    cypress_version VARCHAR(64) NOT NULL,
    browser_name    VARCHAR(64) NOT NULL,
    browser_version VARCHAR(64) NOT NULL,
    headless    BOOLEAN DEFAULT TRUE,
    os_name	    VARCHAR(64) NOT NULL,
    os_version  VARCHAR(64) NOT NULL,
    node_version    VARCHAR(64) NOT NULL,
    create_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
    update_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000)
);

CREATE TABLE IF NOT EXISTS {{.prefix}}spec_executions (
    id          CHAR(26) PRIMARY KEY,
    file	    TEXT NOT NULL,
    server	    TEXT DEFAULT NULL,
    state	    VARCHAR(64) DEFAULT NULL,
    duration    BIGINT NOT NULL DEFAULT 0,
    tests       SMALLINT NOT NULL DEFAULT 0,
    pass        SMALLINT NOT NULL DEFAULT 0,
    fail        SMALLINT NOT NULL DEFAULT 0,
    pending     SMALLINT NOT NULL DEFAULT 0,
    skipped     SMALLINT NOT NULL DEFAULT 0,
    sort_weight SMALLINT NOT NULL DEFAULT 0,
    test_start_at   BIGINT DEFAULT NULL,
    test_end_at BIGINT DEFAULT NULL,
    create_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
    update_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),

    cycle_id    CHAR(26) REFERENCES {{.prefix}}cycles(id) NOT NULL,

    UNIQUE(file, cycle_id)
);

CREATE TABLE IF NOT EXISTS {{.prefix}}case_executions (
    id          CHAR(26) PRIMARY KEY,
    title	    TEXT[] NOT NULL,
    full_title  TEXT NOT NULL,
    key	        VARCHAR(64) DEFAULT NULL,
    key_step    VARCHAR(64) DEFAULT NULL,
    state	    VARCHAR(64) DEFAULT NULL,
    duration    BIGINT NOT NULL DEFAULT 0,
    test_start_at   BIGINT DEFAULT NULL,
    code        TEXT,
    error_display   TEXT,
    error_frame TEXT,
    screenshot  JSONB,
    create_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),
    update_at   BIGINT NOT NULL DEFAULT (extract(epoch from now()) * 1000),

    cycle_id    CHAR(26) REFERENCES {{.prefix}}cycles(id) NOT NULL,
    spec_execution_id   CHAR(26) REFERENCES {{.prefix}}spec_executions(id) NOT NULL,

    UNIQUE(full_title, cycle_id, spec_execution_id)
);
