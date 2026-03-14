BEGIN;

-- =============================================================================
-- 1. ОЧИСТКА ДАННЫХ (CASCADE удалит всё: таблицы, индексы, триггеры)
-- =============================================================================
DROP TABLE IF EXISTS
    ad_promotions, promotion_services, ad_rentals, reports, reviews,
    messages, chats, favorites, ad_images, advertisements,
    campus_locations, faculties, universities, categories, users
    CASCADE;

DROP TYPE IF EXISTS ad_status, deal_type CASCADE;
DROP FUNCTION IF EXISTS update_user_rating, update_updated_at_column CASCADE;

-- =============================================================================
-- 2. РАСШИРЕНИЯ И СЕРВИСНЫЕ ФУНКЦИИ
-- =============================================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- Для UUID
CREATE EXTENSION IF NOT EXISTS "pg_trgm";   -- Для нечеткого поиска

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ language 'plpgsql';

-- =============================================================================
-- 3. ТАБЛИЦЫ УНИВЕРСИТЕТСКОЙ ИНФРАСТРУКТУРЫ
-- =============================================================================
CREATE TABLE universities (
                              id SERIAL PRIMARY KEY,
                              name VARCHAR(255) NOT NULL,
                              domain VARCHAR(50) UNIQUE NOT NULL, -- например, 'bmstu.ru'
                              city VARCHAR(100) NOT NULL
);

CREATE TABLE faculties (
                           id SERIAL PRIMARY KEY,
                           university_id INTEGER NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
                           name VARCHAR(255) NOT NULL
);

CREATE TABLE campus_locations (
                                  id SERIAL PRIMARY KEY,
                                  university_id INTEGER NOT NULL REFERENCES universities(id) ON DELETE CASCADE,
                                  name VARCHAR(255) NOT NULL, -- 'Общежитие №3', 'УЦ Бауманки'
                                  address TEXT,
                                  geo_point POINT -- Точные координаты для карты
);

-- =============================================================================
-- 4. ПОЛЬЗОВАТЕЛИ (STUDENT PROFILE)
-- =============================================================================
CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       university_id INTEGER REFERENCES universities(id) ON DELETE SET NULL,
                       faculty_id INTEGER REFERENCES faculties(id) ON DELETE SET NULL,

                       email VARCHAR(255) UNIQUE NOT NULL, -- Сюда идет edu_email
                       phone VARCHAR(20) UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       name VARCHAR(100) NOT NULL,
                       avatar_url TEXT,
                       role VARCHAR(20) DEFAULT 'user',
    -- Студенческие данные
                       student_card_id VARCHAR(50),
                       is_verified_student BOOLEAN DEFAULT false,
                       rating DECIMAL(3, 2) DEFAULT 0.00 CHECK (rating >= 0 AND rating <= 5),

                       is_blocked BOOLEAN DEFAULT false,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_users_modtime
    BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 5. КАТЕГОРИИ И ТИПЫ СДЕЛОК
-- =============================================================================
CREATE TABLE categories (
                            id SERIAL PRIMARY KEY,
                            parent_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
                            name VARCHAR(100) NOT NULL,
                            slug VARCHAR(100) UNIQUE NOT NULL,
                            icon_url TEXT
);

CREATE TYPE ad_status AS ENUM ('draft', 'moderating', 'active', 'rejected', 'sold', 'closed');
CREATE TYPE deal_type AS ENUM ('sell', 'rent', 'exchange', 'free'); -- Продажа, Аренда, Обмен, Даром

-- =============================================================================
-- 6. ОБЪЯВЛЕНИЯ (CAMPUS CORE)
-- =============================================================================
CREATE TABLE advertisements (
                                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                category_id INTEGER NOT NULL REFERENCES categories(id),
                                university_id INTEGER REFERENCES universities(id) ON DELETE CASCADE,
                                location_id INTEGER REFERENCES campus_locations(id) ON DELETE SET NULL,

                                title VARCHAR(255) NOT NULL,
                                description TEXT NOT NULL,
                                price NUMERIC(15, 2) NOT NULL DEFAULT 0 CHECK (price >= 0),
                                currency VARCHAR(3) DEFAULT 'RUB',

                                status ad_status DEFAULT 'moderating',
                                deal_type deal_type DEFAULT 'sell',

    -- Гибкие свойства (ISBN для книг, курс, автор, модель техники)
                                extra_props JSONB DEFAULT '{}'::jsonb,

                                views_count INTEGER DEFAULT 0,
                                rejection_reason TEXT, -- Для модерации

    -- Полнотекстовый поиск
                                search_vector tsvector GENERATED ALWAYS AS (
                                    to_tsvector('russian', title || ' ' || description)
                                    ) STORED,

                                expires_at TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP + interval '30 days'),
                                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_ads_modtime
    BEFORE UPDATE ON advertisements FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 7. ФОТО, ИЗБРАННОЕ И ЖАЛОБЫ
-- =============================================================================
CREATE TABLE ad_images (
                           id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                           ad_id UUID NOT NULL REFERENCES advertisements(id) ON DELETE CASCADE,
                           image_url TEXT NOT NULL,
                           is_main BOOLEAN DEFAULT false,
                           priority INTEGER DEFAULT 0
);

CREATE TABLE favorites (
                           user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                           ad_id UUID REFERENCES advertisements(id) ON DELETE CASCADE,
                           created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                           PRIMARY KEY (user_id, ad_id)
);

CREATE TABLE reports (
                         id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                         reporter_id UUID REFERENCES users(id) ON DELETE SET NULL,
                         ad_id UUID REFERENCES advertisements(id) ON DELETE CASCADE,
                         reason TEXT NOT NULL,
                         status VARCHAR(20) DEFAULT 'new',
                         created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 8. МЕССЕНДЖЕР И ОТЗЫВЫ
-- =============================================================================
CREATE TABLE chats (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       ad_id UUID REFERENCES advertisements(id) ON DELETE SET NULL,
                       buyer_id UUID REFERENCES users(id) ON DELETE CASCADE,
                       seller_id UUID REFERENCES users(id) ON DELETE CASCADE,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       UNIQUE(ad_id, buyer_id, seller_id)
);

CREATE TABLE messages (
                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                          chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
                          sender_id UUID NOT NULL REFERENCES users(id),
                          content TEXT NOT NULL,
                          is_read BOOLEAN DEFAULT false,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE reviews (
                         id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                         ad_id UUID REFERENCES advertisements(id) ON DELETE SET NULL,
                         from_user_id UUID NOT NULL REFERENCES users(id),
                         to_user_id UUID NOT NULL REFERENCES users(id),
                         rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
                         comment TEXT,
                         created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                         UNIQUE(from_user_id, ad_id)
);

-- =============================================================================
-- 9. АРЕНДА И ШЕРИНГ (Для аренды учебников/техники)
-- =============================================================================
CREATE TABLE ad_rentals (
                            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                            ad_id UUID NOT NULL REFERENCES advertisements(id) ON DELETE CASCADE,
                            tenant_id UUID NOT NULL REFERENCES users(id),
                            start_date TIMESTAMP WITH TIME ZONE NOT NULL,
                            end_date TIMESTAMP WITH TIME ZONE NOT NULL,
                            status VARCHAR(20) DEFAULT 'requested', -- requested, active, completed, cancelled
                            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 10. МОНЕТИЗАЦИЯ (Продвижение объявлений в кампусе)
-- =============================================================================
CREATE TABLE promotion_services (
                                    id SERIAL PRIMARY KEY,
                                    name VARCHAR(100) NOT NULL,
                                    price NUMERIC(10, 2) NOT NULL
);

CREATE TABLE ad_promotions (
                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                               ad_id UUID NOT NULL REFERENCES advertisements(id) ON DELETE CASCADE,
                               service_id INTEGER NOT NULL REFERENCES promotion_services(id),
                               end_date TIMESTAMP WITH TIME ZONE NOT NULL,
                               is_active BOOLEAN DEFAULT true
);

-- =============================================================================
-- 11. ЛОГИКА РЕЙТИНГА (Trigger)
-- =============================================================================
CREATE OR REPLACE FUNCTION update_user_rating()
RETURNS TRIGGER AS $$
BEGIN
UPDATE users
SET rating = (SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE to_user_id = COALESCE(NEW.to_user_id, OLD.to_user_id))
WHERE id = COALESCE(NEW.to_user_id, OLD.to_user_id);
RETURN COALESCE(NEW, OLD);
END;
$$ language 'plpgsql';

CREATE TRIGGER trg_after_review_insert
    AFTER INSERT OR UPDATE OR DELETE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_user_rating();

-- =============================================================================
-- 12. ИНДЕКСЫ ДЛЯ HIGHLOAD И ПОИСКА
-- =============================================================================
CREATE INDEX idx_ads_search ON advertisements USING GIN(search_vector);
CREATE INDEX idx_ads_props ON advertisements USING GIN(extra_props);
CREATE INDEX idx_ads_university ON advertisements(university_id);
CREATE INDEX idx_ads_location ON advertisements(location_id);
CREATE INDEX idx_ads_status_active ON advertisements(status) WHERE status = 'active';
CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);
CREATE INDEX idx_users_uni ON users(university_id);

-- ==========================================
-- ТЕСТОВЫЕ ДАННЫЕ (Наполнение справочников)
-- ==========================================

-- 1. Добавляем ВУЗ
INSERT INTO universities (id, name, domain, city)
VALUES (1, 'МГТУ им. Н.Э. Баумана', 'bmstu.ru', 'Москва')
    ON CONFLICT DO NOTHING;

-- 2. Добавляем категории
INSERT INTO categories (id, name, slug) VALUES
                                            (1, 'Учебники и книги', 'books'),
                                            (2, 'Электроника', 'electronics'),
                                            (3, 'Бытовая техника', 'appliances')
    ON CONFLICT DO NOTHING;

-- 3. Добавляем АДМИНА
-- Пароль для входа: admin123
-- Хеш сгенерирован через bcrypt (DefaultCost)
INSERT INTO users (id, email, phone, password_hash, name, role, university_id, is_verified_student)
VALUES (
           uuid_generate_v4(),
           'admin@campus.ru',
           '+70000000000',
           '$2a$10$XmO9K17vM88yS6.S9O9MSe7I8h8E/kH6vMvV6Z7n7.hGjD6gY9.V6v',
           'Главный Админ',
           'admin',
           1,
           true
       )
    ON CONFLICT DO NOTHING;
COMMIT;