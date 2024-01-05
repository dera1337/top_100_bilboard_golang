CREATE TABLE IF NOT EXISTS public.song_information
(
    song_information_id serial NOT NULL,
    title character varying NOT NULL,
    author character varying NOT NULL,
    youtube_url character varying NOT NULL,
    image_url character varying NOT NULL,
    current_rank_number int NOT NULL,
    previous_rank_number int NULL 
);

CREATE TABLE IF NOT EXISTS public.user
(
    fcm_token character varying NOT NULL PRIMARY KEY,
    is_premium BOOLEAN NOT NULL
);