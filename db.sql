--
-- PostgreSQL database dump
--

-- Dumped from database version 14.1
-- Dumped by pg_dump version 14.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: accounts; Type: TABLE; Schema: public; Owner: kvell
--

CREATE TABLE public.accounts (
    id integer NOT NULL,
    is_enabled boolean DEFAULT true NOT NULL,
    is_test boolean DEFAULT false NOT NULL,
    rebill_enabled boolean DEFAULT false NOT NULL,
    refund_enabled boolean DEFAULT true NOT NULL,
    reversal_enabled boolean DEFAULT true NOT NULL,
    partial_confirm_enabled boolean DEFAULT false NOT NULL,
    partial_reversal_enabled boolean DEFAULT false NOT NULL,
    partial_refund_enabled boolean DEFAULT false NOT NULL,
    currency_conversion_enabled boolean DEFAULT false NOT NULL,
    settings jsonb DEFAULT '{}'::jsonb NOT NULL,
    currency_id integer NOT NULL,
    channel_id integer NOT NULL
);


ALTER TABLE public.accounts OWNER TO kvell;

--
-- Name: accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: kvell
--

CREATE SEQUENCE public.accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.accounts_id_seq OWNER TO kvell;

--
-- Name: accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: kvell
--

ALTER SEQUENCE public.accounts_id_seq OWNED BY public.accounts.id;


--
-- Name: channels; Type: TABLE; Schema: public; Owner: kvell
--

CREATE TABLE public.channels (
    id integer NOT NULL,
    type_id integer NOT NULL,
    key character varying(255) NOT NULL
);


ALTER TABLE public.channels OWNER TO kvell;

--
-- Name: currencies; Type: TABLE; Schema: public; Owner: kvell
--

CREATE TABLE public.currencies (
    id integer NOT NULL,
    numeric_code integer NOT NULL,
    name character varying(255) NOT NULL,
    char_code character varying(3) NOT NULL,
    exponent integer
);


ALTER TABLE public.currencies OWNER TO kvell;

--
-- Name: currencies_id_seq; Type: SEQUENCE; Schema: public; Owner: kvell
--

CREATE SEQUENCE public.currencies_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.currencies_id_seq OWNER TO kvell;

--
-- Name: currencies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: kvell
--

ALTER SEQUENCE public.currencies_id_seq OWNED BY public.currencies.id;


--
-- Name: instruments; Type: TABLE; Schema: public; Owner: kvell
--

CREATE TABLE public.instruments (
    id integer NOT NULL,
    key character varying(255) NOT NULL
);


ALTER TABLE public.instruments OWNER TO kvell;

--
-- Name: routers; Type: TABLE; Schema: public; Owner: kvell
--

CREATE TABLE public.routers (
    id integer NOT NULL,
    key character varying(255) NOT NULL
);


ALTER TABLE public.routers OWNER TO kvell;

--
-- Name: routes; Type: TABLE; Schema: public; Owner: kvell
--

CREATE TABLE public.routes (
    id integer NOT NULL,
    profile_id integer NOT NULL,
    instrument_id integer NOT NULL,
    account_id integer,
    router_id integer,
    settings jsonb
);


ALTER TABLE public.routes OWNER TO kvell;

--
-- Name: routes_id_seq; Type: SEQUENCE; Schema: public; Owner: kvell
--

CREATE SEQUENCE public.routes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.routes_id_seq OWNER TO kvell;

--
-- Name: routes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: kvell
--

ALTER SEQUENCE public.routes_id_seq OWNED BY public.routes.id;


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: kvell
--

CREATE TABLE public.transactions (
    id integer NOT NULL,
    created timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    type character varying(255) NOT NULL,
    status character varying(255) NOT NULL,
    profile_id integer NOT NULL,
    account_id integer NOT NULL,
    instrument_id integer NOT NULL,
    instrument integer NOT NULL,
    amount integer NOT NULL,
    currency_id integer NOT NULL,
    amount_converted integer NOT NULL,
    currency_converted_id integer NOT NULL,
    authcode character varying(8),
    rrn character varying(255),
    response_code character varying(255),
    remote_id character varying(255),
    order_id character varying(255) NOT NULL,
    reference_id integer,
    threedsecure10 jsonb,
    threedsecure20 jsonb,
    threedsmethodurl jsonb
);


ALTER TABLE public.transactions OWNER TO kvell;

--
-- Name: transactions_id_seq; Type: SEQUENCE; Schema: public; Owner: kvell
--

CREATE SEQUENCE public.transactions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.transactions_id_seq OWNER TO kvell;

--
-- Name: transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: kvell
--

ALTER SEQUENCE public.transactions_id_seq OWNED BY public.transactions.id;


--
-- Name: accounts id; Type: DEFAULT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.accounts ALTER COLUMN id SET DEFAULT nextval('public.accounts_id_seq'::regclass);


--
-- Name: currencies id; Type: DEFAULT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.currencies ALTER COLUMN id SET DEFAULT nextval('public.currencies_id_seq'::regclass);


--
-- Name: routes id; Type: DEFAULT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.routes ALTER COLUMN id SET DEFAULT nextval('public.routes_id_seq'::regclass);


--
-- Name: transactions id; Type: DEFAULT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.transactions ALTER COLUMN id SET DEFAULT nextval('public.transactions_id_seq'::regclass);


--
-- Name: accounts accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- Name: channels channel_key_uix; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.channels
    ADD CONSTRAINT channel_key_uix UNIQUE (key);


--
-- Name: channels channels_pkey; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.channels
    ADD CONSTRAINT channels_pkey PRIMARY KEY (id);


--
-- Name: currencies currencies_pkey; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.currencies
    ADD CONSTRAINT currencies_pkey PRIMARY KEY (id);


--
-- Name: currencies currency_numeric_code_uix; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.currencies
    ADD CONSTRAINT currency_numeric_code_uix UNIQUE (numeric_code);


--
-- Name: instruments instruments_key_key; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.instruments
    ADD CONSTRAINT instruments_key_key UNIQUE (key);


--
-- Name: instruments instruments_pkey; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.instruments
    ADD CONSTRAINT instruments_pkey PRIMARY KEY (id);


--
-- Name: routers routers_key_key; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.routers
    ADD CONSTRAINT routers_key_key UNIQUE (key);


--
-- Name: routers routers_pkey; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.routers
    ADD CONSTRAINT routers_pkey PRIMARY KEY (id);


--
-- Name: routes routes_pkey; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.routes
    ADD CONSTRAINT routes_pkey PRIMARY KEY (id);


--
-- Name: routes routes_profile_id_instrument_id_key; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.routes
    ADD CONSTRAINT routes_profile_id_instrument_id_key UNIQUE (profile_id, instrument_id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: type_id_idx; Type: INDEX; Schema: public; Owner: kvell
--

CREATE INDEX type_id_idx ON public.channels USING btree (type_id);


--
-- Name: accounts accounts_channel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.channels(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: accounts accounts_currency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_currency_id_fkey FOREIGN KEY (currency_id) REFERENCES public.currencies(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: routes routes_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.routes
    ADD CONSTRAINT routes_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: routes routes_instrument_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.routes
    ADD CONSTRAINT routes_instrument_id_fkey FOREIGN KEY (instrument_id) REFERENCES public.instruments(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: routes routes_router_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.routes
    ADD CONSTRAINT routes_router_id_fkey FOREIGN KEY (router_id) REFERENCES public.routers(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: transactions transactions_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: transactions transactions_currency_converted_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_currency_converted_id_fkey FOREIGN KEY (currency_converted_id) REFERENCES public.currencies(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: transactions transactions_currency_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_currency_id_fkey FOREIGN KEY (currency_id) REFERENCES public.currencies(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: transactions transactions_instrument_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_instrument_id_fkey FOREIGN KEY (instrument_id) REFERENCES public.instruments(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: transactions transactions_reference_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: kvell
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_reference_id_fkey FOREIGN KEY (reference_id) REFERENCES public.transactions(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--

