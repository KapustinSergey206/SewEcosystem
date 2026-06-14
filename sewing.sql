--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5
-- Dumped by pg_dump version 17.5

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
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
-- Name: company_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.company_info (
    id bigint NOT NULL,
    name text NOT NULL,
    address text NOT NULL,
    phone text DEFAULT ''::text NOT NULL,
    lat numeric(10,6),
    lon numeric(10,6)
);

ALTER TABLE public.company_info OWNER TO postgres;

--
-- Name: company_info_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.company_info_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.company_info_id_seq OWNER TO postgres;

ALTER SEQUENCE public.company_info_id_seq OWNED BY public.company_info.id;

--
-- Name: order_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_items (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL,
    unit_price numeric(12,2) NOT NULL,
    subtotal numeric(12,2) NOT NULL,
    CONSTRAINT order_items_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT order_items_subtotal_check CHECK ((subtotal >= (0)::numeric)),
    CONSTRAINT order_items_unit_price_check CHECK ((unit_price >= (0)::numeric))
);

ALTER TABLE public.order_items OWNER TO postgres;

--
-- Name: order_items_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.order_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.order_items_id_seq OWNER TO postgres;

ALTER SEQUENCE public.order_items_id_seq OWNED BY public.order_items.id;

--
-- Name: order_status_history; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_status_history (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    status text NOT NULL,
    comment text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE public.order_status_history OWNER TO postgres;

--
-- Name: order_status_history_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.order_status_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.order_status_history_id_seq OWNER TO postgres;

ALTER SEQUENCE public.order_status_history_id_seq OWNED BY public.order_status_history.id;

--
-- Name: orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.orders (
    id bigint NOT NULL,
    order_number text GENERATED ALWAYS AS (('ORDER-'::text || lpad((id)::text, 6, '0'::text))) STORED,
    user_id bigint,
    contact_name text DEFAULT ''::text NOT NULL,
    contact_phone text DEFAULT ''::text NOT NULL,
    status text DEFAULT 'new'::text NOT NULL,
    total_amount numeric(12,2) DEFAULT 0 NOT NULL,
    comment text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT orders_status_check CHECK ((status = ANY (ARRAY['new'::text, 'processing'::text, 'ready'::text, 'shipped'::text, 'completed'::text, 'cancelled'::text])))
);

ALTER TABLE public.orders OWNER TO postgres;

--
-- Name: orders_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.orders_id_seq OWNER TO postgres;

ALTER SEQUENCE public.orders_id_seq OWNED BY public.orders.id;

--
-- Name: passnohash; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.passnohash (
    login character varying(255),
    password character varying(255)
);

ALTER TABLE public.passnohash OWNER TO postgres;

--
-- Name: products; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.products (
    id bigint NOT NULL,
    sku text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    category text DEFAULT 'Спецодежда'::text NOT NULL,
    price numeric(12,2) NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    image_path text DEFAULT ''::text NOT NULL,
    image_path_2 text,
    image_path_3 text,
    CONSTRAINT products_price_check CHECK ((price >= (0)::numeric))
);

ALTER TABLE public.products OWNER TO postgres;

--
-- Name: products_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.products_id_seq OWNER TO postgres;

ALTER SEQUENCE public.products_id_seq OWNED BY public.products.id;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    full_name text DEFAULT ''::text NOT NULL,
    email text NOT NULL,
    phone text DEFAULT ''::text NOT NULL,
    password_hash text NOT NULL,
    role text DEFAULT 'client'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_role_check CHECK ((role = ANY (ARRAY['admin'::text, 'client'::text])))
);

ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

--
-- Name: company_info id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.company_info ALTER COLUMN id SET DEFAULT nextval('public.company_info_id_seq'::regclass);

--
-- Name: order_items id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items ALTER COLUMN id SET DEFAULT nextval('public.order_items_id_seq'::regclass);

--
-- Name: order_status_history id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_status_history ALTER COLUMN id SET DEFAULT nextval('public.order_status_history_id_seq'::regclass);

--
-- Name: orders id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders ALTER COLUMN id SET DEFAULT nextval('public.orders_id_seq'::regclass);

--
-- Name: products id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);

--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

--
-- Data for Name: company_info; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.company_info (id, name, address, phone, lat, lon) FROM stdin;
1	Партнёр	г. Фурманов, ул. Социалистический Посёлок 4, 1	+7 (963) 152-96-07	56.129057	47.251026
\.

--
-- Data for Name: order_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_items (id, order_id, product_id, quantity, unit_price, subtotal) FROM stdin;
28	19	1	1	2700.00	2700.00
\.

--
-- Data for Name: order_status_history; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_status_history (id, order_id, status, comment, created_at) FROM stdin;
29	19	new	Заказ создан	2026-06-07 13:39:46.777975+03
\.

--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.orders (id, user_id, contact_name, contact_phone, status, total_amount, comment, created_at, updated_at) FROM stdin;
19	15	Сергей	+79320123212	new	2700.00	32131	2026-06-07 13:39:46.777975+03	2026-06-07 13:39:46.777975+03
\.

--
-- Data for Name: passnohash; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.passnohash (login, password) FROM stdin;
adimn@gmail.com	123456
admin@fewfwefw	4324324
\.

--
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.products (id, sku, name, description, category, price, is_active, created_at, image_path, image_path_2, image_path_3) FROM stdin;
2	LEGION-001	Костюм Легион	Универсальный рабочий костюм с контрастной строчкой и светоотражающими элементами.	Костюмы	1980.00	t	2026-04-07 12:47:27.968392+03	/static/assets/images/gallery/legion/IMG_5451.JPG	/static/assets/images/gallery/legion/IMG_5450.JPG	/static/assets/images/gallery/legion/IMG_5449.JPG
3	GUARD-001	Костюм Охрана	Комплект для охранных структур и постовой службы. Комфортная посадка и износостойкая ткань.	Костюмы	2450.00	t	2026-04-07 12:47:27.968392+03	/static/assets/images/gallery/guard/IMG_5472.JPG	/static/assets/images/gallery/guard/IMG_5470.JPG	/static/assets/images/gallery/guard/IMG_5471.JPG
4	TECH-001	Костюм Техник	Рабочая форма для сервисных специалистов, монтажников и техников.	Костюмы	2300.00	t	2026-04-07 12:47:27.968392+03	/static/assets/images/gallery/technician/IMG_5452.JPG	/static/assets/images/gallery/technician/IMG_5453.JPG	/static/assets/images/gallery/technician/IMG_5454.JPG
6	FLEECE-001	Флисовая куртка	Тёплая корпоративная флисовая куртка для межсезонья и холодных помещений.	Куртки	1850.00	t	2026-04-07 12:47:27.968392+03	/static/assets/images/gallery/jacket-foremen/IMG_5518.JPG	/static/assets/images/gallery/jacket-foremen/IMG_5519.JPG	/static/assets/images/gallery/jacket-foremen/IMG_5519.JPG
7	CORNFLOWER-001	Костюм Василёк	Классический рабочий костюм для производства и сервиса.	Костюмы	2100.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/cornflower/IMG_5476.JPG	/static/assets/images/gallery/cornflower/IMG_5477.JPG	/static/assets/images/gallery/cornflower/IMG_5478.JPG
9	CORNFLOWER-BLUE-001	Костюм Василёк синий	Износостойкий костюм для цеха, склада и выездных работ.	Костюмы	2250.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/cornflower-blue/IMG_5483.JPG	/static/assets/images/gallery/cornflower-blue/IMG_5482.JPG	/static/assets/images/gallery/cornflower-blue/IMG_5484.JPG
27	Trophy-001	Костюм Трофи	Рабочий костюм для автомехаников.	Костюмы	2700.00	t	2026-04-07 13:29:48.526277+03	/static/assets/images/gallery/trophy/trophy.png	/static/assets/images/gallery/trophy/trophy2.png	/static/assets/images/gallery/trophy/trophy3.png
1	AUTOSERVICE-001	Костюм Автосервис	Практичный комплект для сотрудников автосервиса. Потайные застёжки, усиленные швы, удобный крой.	Костюмы	2700.00	t	2026-04-07 12:47:27.968392+03	/static/assets/images/gallery/autoser/IMG_5501.JPG	/static/assets/images/gallery/autoser/IMG_5500.JPG	/static/assets/images/gallery/autoser/IMG_5499.JPG
10	MOUNTAIN-001	Костюм Горный	Практичный комплект для сложных условий эксплуатации.	Костюмы	2950.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/mountain/IMG_5437.JPG	/static/assets/images/gallery/mountain/IMG_5435.JPG	/static/assets/images/gallery/mountain/IMG_5436.JPG
11	FAVORITE-001	Костюм Фаворит	Универсальная модель для повседневной рабочей формы.	Костюмы	2380.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/favorite/IMG_5446.JPG	/static/assets/images/gallery/favorite/IMG_5445.JPG	/static/assets/images/gallery/favorite/IMG_5448.JPG
12	WELDER-001	Костюм Сварщика	Спецодежда для сварочных и производственных работ.	Спецодежда	3200.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/welder/IMG_5506.JPG	/static/assets/images/gallery/welder/IMG_5505.JPG	/static/assets/images/gallery/welder/IMG_5507.JPG
14	GUARD-WARM-001	Костюм Охрана утеплённый	Тёплая форма для охраны и работы на улице.	Утеплённая одежда	3650.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/guard-warm/IMG_5515.JPG	/static/assets/images/gallery/guard-warm/IMG_5516.JPG	/static/assets/images/gallery/guard-warm/IMG_5517.JPG
16	TROUSERS-WARM-001	Брюки утеплённые	Тёплые рабочие брюки для холодного сезона.	Брюки	1750.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/trousers-warm/IMG_5508.JPG	/static/assets/images/gallery/trousers-warm/IMG_5509.JPG	/static/assets/images/gallery/trousers-warm/IMG_5509.JPG
17	WAISTCOAT-WARM-001	Жилет утеплённый	Утеплённый жилет для склада, логистики и производства.	Жилеты	1600.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/waistcoat-warm/IMG_5488.JPG	/static/assets/images/gallery/waistcoat-warm/IMG_5489.JPG	/static/assets/images/gallery/waistcoat-warm/IMG_5485.JPG
18	JACKET-GUARD-001	Куртка охранника	Верхняя одежда для службы охраны и патрулирования.	Куртки	2800.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/jacket-guard/IMG_6525.jpg	/static/assets/images/gallery/jacket-guard/IMG_6526.jpg	/static/assets/images/gallery/jacket-guard/IMG_6526.jpg
20	SPECIAL-POLICE-001	Костюм спецслужб	Тактический комплект для специальных задач.	Спецодежда	3900.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/special-police/IMG_5439.JPG	/static/assets/images/gallery/special-police/IMG_5438.JPG	/static/assets/images/gallery/special-police/IMG_5441.JPG
21	SIGNAL-001	Сигнальный костюм	Рабочая форма повышенной видимости.	Сигнальная одежда	2500.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/signal/IMG_5442.JPG	/static/assets/images/gallery/signal/IMG_5443.JPG	/static/assets/images/gallery/signal/IMG_5444.JPG
22	ROBE-001	Рабочий халат	Халат для персонала производства, склада и сервиса.	Халаты	1450.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/robe/IMG_5469.JPG	/static/assets/images/gallery/robe/IMG_5468.JPG	/static/assets/images/gallery/robe/IMG_5468.JPG
25	Energy-001	Костюм Энергия	Рабочий костюм для электриков.	Костюмы	3000.00	t	2026-04-07 13:27:36.084804+03	/static/assets/images/gallery/energy/energy.jpg	/static/assets/images/gallery/energy/energy.jpg	/static/assets/images/gallery/energy/energy.jpg
26	Ritm-001	Костюм Ритм	Рабочий костюм для автомехаников.	Костюмы	2700.00	t	2026-04-07 13:28:38.113783+03	/static/assets/images/gallery/rhythm/ritm.jpg	/static/assets/images/gallery/rhythm/ritm2.jpg	/static/assets/images/gallery/rhythm/ritm3.jpg
5	VEST-001	Сигнальный жилет	Яркий сигнальный жилет для сотрудников склада, стройки и логистики.	Жилеты	790.00	t	2026-04-07 12:47:27.968392+03	/static/assets/images/gallery/waistcoat-signal/IMG_5492.JPG	/static/assets/images/gallery/waistcoat-signal/IMG_5491.JPG	/static/assets/images/gallery/waistcoat-signal/IMG_5491.JPG
8	CORNFLOWER-RED-001	Костюм Василёк красный	Яркий вариант рабочего костюма с контрастной отделкой.	Костюмы	2250.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/cornflower-red/IMG_5461.JPG	/static/assets/images/gallery/cornflower-red/IMG_5459.JPG	/static/assets/images/gallery/cornflower-red/IMG_5460.JPG
13	STORM-001	Костюм Шторм	Плотная защита для уличных и монтажных работ.	Костюмы	3400.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/storm/IMG_5514.JPG	/static/assets/images/gallery/storm/IMG_5512.JPG	/static/assets/images/gallery/storm/IMG_5513.JPG
15	SIGNAL-WARM-001	Костюм Сигнал утеплённый	Тёплый сигнальный комплект со светоотражающими вставками.	Утеплённая одежда	3550.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/signal-warm/IMG_5503.JPG	/static/assets/images/gallery/signal-warm/IMG_5502.JPG	/static/assets/images/gallery/signal-warm/IMG_5504.JPG
19	ANTI-001	Антистатический костюм	Одежда для специальных производственных условий.	Спецодежда	3300.00	t	2026-04-07 13:10:24.012138+03	/static/assets/images/gallery/anti/IMG_5431.JPG	/static/assets/images/gallery/anti/IMG_5432.JPG	/static/assets/images/gallery/anti/IMG_5433.JPG
23	CAR-WASHER-001	Костюм Автомойка	Рабочий костюм для сотрудников автомойки и сервисных зон.	Костюмы	2350.00	t	2026-04-07 13:25:25.202946+03	/static/assets/images/gallery/car-washer/car-washer.jpg	/static/assets/images/gallery/car-washer/car-washer2.jpg	/static/assets/images/gallery/car-washer/car-washer2.jpg
28	Flis-jacket-001	Флисовая куртка	Флисовая куртка.	Костюмы	1350.00	t	2026-04-07 13:31:44.117286+03	/static/assets/images/gallery/jacket/flis-jacket.jpg	/static/assets/images/gallery/jacket/flis-jacket2.png	/static/assets/images/gallery/jacket/flis-jacket2.png
\.

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, full_name, email, phone, password_hash, role, created_at) FROM stdin;
1	Администратор	adimn@gmail.com	+79201652165	$2a$10$dHhMBRs4ux8bFHmUXcsSKe03bWAHSR/Id3mWMnTplAaGCqmXwpJry	admin	2026-06-07 00:00:00+03
15	Баранов Дмитрий Иванович	gerre@fsdfsd	0978967967	$2a$10$.dvRSLn9xNe/SfhQPyrWYuVfetTJ2zrXODvECAmQdfoagto/lu7R.	client	2026-06-07 13:39:37.620528+03
\.

--
-- Name: company_info_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.company_info_id_seq', 1, true);

--
-- Name: order_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_items_id_seq', 29, true);

--
-- Name: order_status_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_status_history_id_seq', 30, true);

--
-- Name: orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.orders_id_seq', 20, true);

--
-- Name: products_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.products_id_seq', 28, true);

--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 16, true);

--
-- Name: company_info company_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.company_info
    ADD CONSTRAINT company_info_pkey PRIMARY KEY (id);

--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);

--
-- Name: order_status_history order_status_history_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_status_history
    ADD CONSTRAINT order_status_history_pkey PRIMARY KEY (id);

--
-- Name: orders orders_order_number_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_order_number_key UNIQUE (order_number);

--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

--
-- Name: products products_sku_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_sku_key UNIQUE (sku);

--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

--
-- Name: idx_orders_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_created_at ON public.orders USING btree (created_at);

--
-- Name: idx_orders_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_status ON public.orders USING btree (status);

--
-- Name: idx_orders_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_orders_user_id ON public.orders USING btree (user_id);

--
-- Name: idx_order_items_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_order_items_order_id ON public.order_items USING btree (order_id);

--
-- Name: idx_order_items_product_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_order_items_product_id ON public.order_items USING btree (product_id);

--
-- Name: idx_order_status_history_order_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_order_status_history_order_id ON public.order_status_history USING btree (order_id);

--
-- Name: idx_products_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_products_category ON public.products USING btree (category);

--
-- Name: idx_products_is_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_products_is_active ON public.products USING btree (is_active);

--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email ON public.users USING btree (email);

--
-- Name: idx_users_role; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_role ON public.users USING btree (role);

--
-- Name: order_items order_items_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;

--
-- Name: order_items order_items_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id);

--
-- Name: order_status_history order_status_history_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_status_history
    ADD CONSTRAINT order_status_history_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;

--
-- Name: orders orders_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

--
-- PostgreSQL database dump complete
--