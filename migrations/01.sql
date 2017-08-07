--
-- PostgreSQL database dump
--

CREATE DATABASE newshound
    WITH 
    OWNER = newshound
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TABLESPACE = pg_default
    CONNECTION LIMIT = -1;

-- Dumped from database version 9.6.3
-- Dumped by pg_dump version 9.6.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: newshound; Type: SCHEMA; Schema: -; Owner: newshound
--

CREATE SCHEMA newshound;


ALTER SCHEMA newshound OWNER TO postgres;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = newshound, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: alert; Type: TABLE; Schema: newshound; Owner: newshound
--

CREATE TABLE alert (
    id bigint NOT NULL,
    sender_id integer NOT NULL,
    url text,
    "timestamp" timestamp with time zone NOT NULL,
    top_phrases text[],
    top_sentence integer NOT NULL,
    subject text,
    raw_body text,
    body text
);


ALTER TABLE alert OWNER TO postgres;

--
-- Name: alert_id_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE alert_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE alert_id_seq OWNER TO postgres;

--
-- Name: alert_id_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE alert_id_seq OWNED BY alert.id;


--
-- Name: alert_sender_id_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE alert_sender_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE alert_sender_id_seq OWNER TO postgres;

--
-- Name: alert_sender_id_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE alert_sender_id_seq OWNED BY alert.sender_id;


--
-- Name: alert_top_sentence_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE alert_top_sentence_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE alert_top_sentence_seq OWNER TO postgres;

--
-- Name: alert_top_sentence_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE alert_top_sentence_seq OWNED BY alert.top_sentence;


--
-- Name: event; Type: TABLE; Schema: newshound; Owner: newshound
--

CREATE TABLE event (
    id bigint NOT NULL,
    top_phrases text[],
    start timestamp with time zone NOT NULL,
    "end" timestamp with time zone NOT NULL,
    top_sentence bigint NOT NULL,
    top_sender integer NOT NULL
);


ALTER TABLE event OWNER TO postgres;

--
-- Name: event_alert; Type: TABLE; Schema: newshound; Owner: newshound
--

CREATE TABLE event_alert (
    alert_id bigint NOT NULL,
    event_id bigint NOT NULL
);


ALTER TABLE event_alert OWNER TO postgres;

--
-- Name: event_alert_alert_id_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE event_alert_alert_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE event_alert_alert_id_seq OWNER TO postgres;

--
-- Name: event_alert_alert_id_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE event_alert_alert_id_seq OWNED BY event_alert.alert_id;


--
-- Name: event_alert_event_id_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE event_alert_event_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE event_alert_event_id_seq OWNER TO postgres;

--
-- Name: event_alert_event_id_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE event_alert_event_id_seq OWNED BY event_alert.event_id;


--
-- Name: event_id_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE event_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE event_id_seq OWNER TO postgres;

--
-- Name: event_id_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE event_id_seq OWNED BY event.id;


--
-- Name: event_top_sender_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE event_top_sender_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE event_top_sender_seq OWNER TO postgres;

--
-- Name: event_top_sender_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE event_top_sender_seq OWNED BY event.top_sender;


--
-- Name: event_top_sentence_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE event_top_sentence_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE event_top_sentence_seq OWNER TO postgres;

--
-- Name: event_top_sentence_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE event_top_sentence_seq OWNED BY event.top_sentence;


--
-- Name: sender; Type: TABLE; Schema: newshound; Owner: newshound
--

CREATE TABLE sender (
    id integer NOT NULL,
    name text,
    url_index integer,
    color character varying(6)
);


ALTER TABLE sender OWNER TO postgres;

--
-- Name: sender_id_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE sender_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sender_id_seq OWNER TO postgres;

--
-- Name: sender_id_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE sender_id_seq OWNED BY sender.id;


--
-- Name: sentence; Type: TABLE; Schema: newshound; Owner: newshound
--

CREATE TABLE sentence (
    text text,
    phrases text[],
    alert_id bigint NOT NULL,
    id bigint NOT NULL
);


ALTER TABLE sentence OWNER TO postgres;

--
-- Name: sentence_alert_id_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE sentence_alert_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sentence_alert_id_seq OWNER TO postgres;

--
-- Name: sentence_alert_id_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE sentence_alert_id_seq OWNED BY sentence.alert_id;


--
-- Name: sentence_id_seq; Type: SEQUENCE; Schema: newshound; Owner: newshound
--

CREATE SEQUENCE sentence_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sentence_id_seq OWNER TO postgres;

--
-- Name: sentence_id_seq; Type: SEQUENCE OWNED BY; Schema: newshound; Owner: newshound
--

ALTER SEQUENCE sentence_id_seq OWNED BY sentence.id;


--
-- Name: alert id; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY alert ALTER COLUMN id SET DEFAULT nextval('alert_id_seq'::regclass);


--
-- Name: alert sender_id; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY alert ALTER COLUMN sender_id SET DEFAULT nextval('alert_sender_id_seq'::regclass);


--
-- Name: alert top_sentence; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY alert ALTER COLUMN top_sentence SET DEFAULT nextval('alert_top_sentence_seq'::regclass);


--
-- Name: event id; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event ALTER COLUMN id SET DEFAULT nextval('event_id_seq'::regclass);


--
-- Name: event top_sentence; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event ALTER COLUMN top_sentence SET DEFAULT nextval('event_top_sentence_seq'::regclass);


--
-- Name: event top_sender; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event ALTER COLUMN top_sender SET DEFAULT nextval('event_top_sender_seq'::regclass);


--
-- Name: event_alert alert_id; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event_alert ALTER COLUMN alert_id SET DEFAULT nextval('event_alert_alert_id_seq'::regclass);


--
-- Name: event_alert event_id; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event_alert ALTER COLUMN event_id SET DEFAULT nextval('event_alert_event_id_seq'::regclass);


--
-- Name: sender id; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY sender ALTER COLUMN id SET DEFAULT nextval('sender_id_seq'::regclass);


--
-- Name: sentence alert_id; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY sentence ALTER COLUMN alert_id SET DEFAULT nextval('sentence_alert_id_seq'::regclass);


--
-- Name: sentence id; Type: DEFAULT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY sentence ALTER COLUMN id SET DEFAULT nextval('sentence_id_seq'::regclass);


--
-- Name: alert alert_pkey; Type: CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY alert
    ADD CONSTRAINT alert_pkey PRIMARY KEY (id);


--
-- Name: event event_pkey; Type: CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event
    ADD CONSTRAINT event_pkey PRIMARY KEY (id);


--
-- Name: sender sender_pkey; Type: CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY sender
    ADD CONSTRAINT sender_pkey PRIMARY KEY (id);


--
-- Name: sentence sentence_pkey; Type: CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY sentence
    ADD CONSTRAINT sentence_pkey PRIMARY KEY (id);


--
-- Name: fki_fk_alert; Type: INDEX; Schema: newshound; Owner: newshound
--

CREATE INDEX fki_fk_alert ON sentence USING btree (alert_id);


--
-- Name: fki_fk_event; Type: INDEX; Schema: newshound; Owner: newshound
--

CREATE INDEX fki_fk_event ON event_alert USING btree (event_id);


--
-- Name: fki_fk_event_alert; Type: INDEX; Schema: newshound; Owner: newshound
--

CREATE INDEX fki_fk_event_alert ON event_alert USING btree (alert_id);


--
-- Name: fki_fk_sender; Type: INDEX; Schema: newshound; Owner: newshound
--

CREATE INDEX fki_fk_sender ON event USING btree (top_sender);


--
-- Name: fki_fk_sentence; Type: INDEX; Schema: newshound; Owner: newshound
--

CREATE INDEX fki_fk_sentence ON event USING btree (top_sentence);


--
-- Name: sentence fk_alert; Type: FK CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY sentence
    ADD CONSTRAINT fk_alert FOREIGN KEY (alert_id) REFERENCES alert(id);


--
-- Name: event_alert fk_alert; Type: FK CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event_alert
    ADD CONSTRAINT fk_alert FOREIGN KEY (alert_id) REFERENCES alert(id);


--
-- Name: event_alert fk_event; Type: FK CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event_alert
    ADD CONSTRAINT fk_event FOREIGN KEY (event_id) REFERENCES event(id);


--
-- Name: alert fk_sender; Type: FK CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY alert
    ADD CONSTRAINT fk_sender FOREIGN KEY (sender_id) REFERENCES sender(id);


--
-- Name: event fk_sender; Type: FK CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event
    ADD CONSTRAINT fk_sender FOREIGN KEY (top_sender) REFERENCES sender(id);


--
-- Name: event fk_sentence; Type: FK CONSTRAINT; Schema: newshound; Owner: newshound
--

ALTER TABLE ONLY event
    ADD CONSTRAINT fk_sentence FOREIGN KEY (top_sentence) REFERENCES sentence(id);


--
-- PostgreSQL database dump complete
--

