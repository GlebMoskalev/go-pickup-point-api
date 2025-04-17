CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE roles_enum AS ENUM('employee', 'moderator');
CREATE TYPE cities_enum AS ENUM('Москва', 'Санкт-Петербург', 'Казань');
CREATE TYPE statuses_enum AS ENUM('in_progress', 'close');
CREATE TYPE product_types_enum AS ENUM('электроника', 'одежда', 'обувь');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role roles_enum NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE pvz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    city cities_enum NOT NULL
);

CREATE TABLE receptions(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    pvz_id UUID NOT NULL REFERENCES pvz(id) ON DELETE CASCADE,
    status statuses_enum NOT NULL
);

CREATE TABLE products(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    type product_types_enum NOT NULL,
    reception_id UUID NOT NULL REFERENCES receptions(id) ON DELETE CASCADE,
    order_number SERIAL
)