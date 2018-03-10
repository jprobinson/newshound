-- Database: newshound

-- DROP DATABASE newshound;

CREATE DATABASE newshound
    WITH 
    OWNER = cloudsqlsuperuser
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF8'
    LC_CTYPE = 'en_US.UTF8'
    TABLESPACE = pg_default
    CONNECTION LIMIT = -1;

-- SCHEMA: newshound

-- DROP SCHEMA newshound ;

CREATE SCHEMA newshound
    AUTHORIZATION newshound;

CREATE SEQUENCE newshound.sender_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 93333333333333333
    CACHE 1;

ALTER SEQUENCE newshound.sender_id_seq
    OWNER TO newshound;

CREATE SEQUENCE newshound.sentence_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 93333333333333333
    CACHE 1;

ALTER SEQUENCE newshound.sentence_id_seq
    OWNER TO newshound;

CREATE SEQUENCE newshound.event_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 93333333333333333
    CACHE 1;

ALTER SEQUENCE newshound.event_id_seq
    OWNER TO newshound;

CREATE SEQUENCE newshound.alert_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 93333333333333333
    CACHE 1;

ALTER SEQUENCE newshound.alert_id_seq
    OWNER TO newshound;


-- Table: newshound.sender

-- DROP TABLE newshound.sender;

CREATE TABLE newshound.sender
(
    id integer NOT NULL DEFAULT nextval('sender_id_seq'::regclass),
    name text COLLATE pg_catalog."default",
    url_index integer,
    color character varying(6) COLLATE pg_catalog."default",
    CONSTRAINT sender_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE newshound.sender
    OWNER to newshound;


-- Table: newshound.alert

-- DROP TABLE newshound.alert;

CREATE TABLE newshound.alert
(
    id integer NOT NULL DEFAULT nextval('alert_id_seq'::regclass),
    sender_id integer NOT NULL,
    url text COLLATE pg_catalog."default",
    "timestamp" timestamp without time zone NOT NULL,
    top_phrases text[] COLLATE pg_catalog."default",
    top_sentence integer,
    subject text COLLATE pg_catalog."default",
    raw_body text COLLATE pg_catalog."default",
    body text COLLATE pg_catalog."default",
    CONSTRAINT alert_pkey PRIMARY KEY (id),
    CONSTRAINT fk_alert_sender FOREIGN KEY (sender_id)
        REFERENCES newshound.sender (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fk_alert_top_sentence FOREIGN KEY (top_sentence)
        REFERENCES newshound.sentence (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE newshound.alert
    OWNER to newshound;

-- Table: newshound.sentence

-- DROP TABLE newshound.sentence;

CREATE TABLE newshound.sentence
(
    text text COLLATE pg_catalog."default",
    phrases text[] COLLATE pg_catalog."default",
    alert_id bigint NOT NULL,
    id bigint NOT NULL DEFAULT nextval('sentence_id_seq'::regclass),
    CONSTRAINT sentence_pkey PRIMARY KEY (id),
    CONSTRAINT fk_sentence_alert FOREIGN KEY (alert_id)
        REFERENCES newshound.alert (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE newshound.sentence
    OWNER to newshound;

-- Table: newshound.event

-- DROP TABLE newshound.event;

CREATE TABLE newshound.event
(
    id integer NOT NULL DEFAULT nextval('event_id_seq'::regclass),
    top_phrases text[] COLLATE pg_catalog."default",
    start timestamp without time zone,
    "end" timestamp without time zone,
    top_sentence integer NOT NULL,
    top_sender integer NOT NULL,
    CONSTRAINT pk_event PRIMARY KEY (id),
    CONSTRAINT fk_event_top_sender FOREIGN KEY (top_sender)
        REFERENCES newshound.sender (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fk_event_top_sentence FOREIGN KEY (top_sentence)
        REFERENCES newshound.sentence (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE newshound.event
    OWNER to newshound;

-- Table: newshound.event_alert

-- DROP TABLE newshound.event_alert;

CREATE TABLE newshound.event_alert
(
    event_id integer NOT NULL,
    alert_id integer NOT NULL,
    CONSTRAINT fk_event_alert_alert FOREIGN KEY (alert_id)
        REFERENCES newshound.alert (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fk_event_alert_event FOREIGN KEY (event_id)
        REFERENCES newshound.event (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE newshound.event_alert
    OWNER to newshound;
