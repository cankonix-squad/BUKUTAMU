-- Data awal bidang (silakan sesuaikan dengan struktur organisasi terbaru)
INSERT INTO bidang (nama) VALUES
    ('Sekretariat'),
    ('Bidang Sumber Daya Air'),
    ('Bidang Jalan dan Jembatan'),
    ('Bidang Drainase'),
    ('Bidang Pemeliharaan'),
    ('UPT / Unit Pelaksana Teknis')
ON CONFLICT (nama) DO NOTHING;
