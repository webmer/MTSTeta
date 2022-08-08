CREATE SCHEMA analytic;

-- Table: analytic.task
-- DROP TABLE analytic.task;
CREATE TABLE IF NOT EXISTS analytic.task
(
    uuid uuid NOT NULL,
    login character varying(50) COLLATE pg_catalog."default",
    date_create timestamp without time zone NOT NULL,
    date_action timestamp without time zone,
    status boolean,
    CONSTRAINT tasks_pkey PRIMARY KEY (uuid)
);

-- Table: analytic.message
-- DROP TABLE analytic.message;
CREATE TABLE IF NOT EXISTS analytic.message
(
    uuid uuid NOT NULL,
    task_uuid uuid NOT NULL,
    date_create timestamp without time zone NOT NULL,
    type character varying(50) COLLATE pg_catalog."default" NOT NULL,
    value character varying(50) COLLATE pg_catalog."default"
);