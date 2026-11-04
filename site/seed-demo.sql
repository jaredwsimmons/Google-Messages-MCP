PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE conversations (
    conversation_id TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    is_group INTEGER NOT NULL DEFAULT 0,
    participants TEXT NOT NULL DEFAULT '[]',
    last_message_ts INTEGER NOT NULL DEFAULT 0,
    unread_count INTEGER NOT NULL DEFAULT 0
);
INSERT INTO conversations VALUES('conv1','Sarah Chen',0,'[{"name":"Sarah Chen","number":"+14155551234"}]',1738958400000,0);
INSERT INTO conversations VALUES('conv2','Marcus Johnson',0,'[{"name":"Marcus Johnson","number":"+12125559876"}]',1738956600000,2);
INSERT INTO conversations VALUES('conv3','Weekend Hiking Group',1,'[{"name":"Emily Park","number":"+13105553456"},{"name":"David Kim","number":"+14085557890"},{"name":"Alex Thompson","number":"+17185552222"}]',1738954800000,0);
INSERT INTO conversations VALUES('conv4','Emily Park',0,'[{"name":"Emily Park","number":"+13105553456"}]',1738951200000,0);
INSERT INTO conversations VALUES('conv5','Lisa Rodriguez',0,'[{"name":"Lisa Rodriguez","number":"+12025551111"}]',1738947600000,1);
INSERT INTO conversations VALUES('conv6','David Kim',0,'[{"name":"David Kim","number":"+14085557890"}]',1738944000000,0);
INSERT INTO conversations VALUES('conv7','Rachel Green',0,'[{"name":"Rachel Green","number":"+16505553333"}]',1738940400000,0);
INSERT INTO conversations VALUES('conv8','Alex Thompson',0,'[{"name":"Alex Thompson","number":"+17185552222"}]',1738936800000,0);
CREATE TABLE messages (
    message_id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL DEFAULT '',
    sender_name TEXT NOT NULL DEFAULT '',
    sender_number TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    timestamp_ms INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT '',
    is_from_me INTEGER NOT NULL DEFAULT 0,
    media_id TEXT NOT NULL DEFAULT '',
    mime_type TEXT NOT NULL DEFAULT '',
    decryption_key TEXT NOT NULL DEFAULT '',
    reactions TEXT NOT NULL DEFAULT '',
    reply_to_id TEXT NOT NULL DEFAULT ''
);
INSERT INTO messages VALUES('m1a','conv1','Sarah Chen','+14155551234','Hey! Are you free for dinner tonight?',1738951200000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m1b','conv1','Me','+15551234567','Yes! What did you have in mind?',1738952100000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m1c','conv1','Sarah Chen','+14155551234','There is a new Thai place on Valencia that just opened. Heard great things about their pad see ew',1738953000000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m1d','conv1','Me','+15551234567','That sounds perfect! What time works for you?',1738954800000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m1e','conv1','Sarah Chen','+14155551234','How about 7:30? I can make a reservation',1738956600000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m1f','conv1','Me','+15551234567','Perfect, see you there!',1738958400000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m2a','conv2','Marcus Johnson','+12125559876','Quick update on the project - we hit our Q1 milestone early!',1738944000000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m2b','conv2','Me','+15551234567','That is awesome news! The team did a great job.',1738945800000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m2c','conv2','Marcus Johnson','+12125559876','Agreed. Want to hop on a call Monday to discuss next steps?',1738947600000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m2d','conv2','Marcus Johnson','+12125559876','Also, I sent over the slide deck to review when you get a chance',1738956600000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m3a','conv3','Emily Park','+13105553456','Who is in for Muir Woods this Saturday? Weather looks amazing',1738940400000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m3b','conv3','David Kim','+14085557890','Count me in! Should we do the Dipsea Trail?',1738942200000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m3c','conv3','Me','+15551234567','I am in! Let us meet at the parking lot at 9am?',1738944000000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m3d','conv3','Alex Thompson','+17185552222','Perfect! I will bring trail mix and water for everyone',1738945800000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m3e','conv3','Emily Park','+13105553456','Amazing! It is going to be a great day. Do not forget sunscreen!',1738954800000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m4a','conv4','Emily Park','+13105553456','Thanks for the book recommendation! I am already halfway through it',1738940400000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m4b','conv4','Me','+15551234567','Glad you are enjoying it! The second half gets even better',1738951200000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m5a','conv5','Lisa Rodriguez','+12025551111','Are we still on for coffee tomorrow morning?',1738936800000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m5b','conv5','Me','+15551234567','Absolutely! Blue Bottle at 10?',1738938600000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m5c','conv5','Lisa Rodriguez','+12025551111','Sounds great! I have some exciting news to share',1738947600000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m6a','conv6','Me','+15551234567','Hey, did you see the Warriors game last night?',1738933200000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m6b','conv6','David Kim','+14085557890','Incredible comeback! Curry was unreal in the 4th quarter',1738936800000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m6c','conv6','Me','+15551234567','We should catch the next home game together',1738944000000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m7a','conv7','Rachel Green','+16505553333','Just landed! Flight was smooth. Thanks for the ride to the airport',1738929600000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m7b','conv7','Me','+15551234567','Anytime! Have an amazing trip',1738940400000,'delivered',1,'','','','','');
INSERT INTO messages VALUES('m8a','conv8','Alex Thompson','+17185552222','Found that restaurant we were talking about - it is called Nopa',1738929600000,'delivered',0,'','','','','');
INSERT INTO messages VALUES('m8b','conv8','Me','+15551234567','Nice find! Let us go next week',1738936800000,'delivered',1,'','','','','');
CREATE TABLE contacts (
    contact_id TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    number TEXT NOT NULL DEFAULT ''
);
INSERT INTO contacts VALUES('c1','Sarah Chen','+14155551234');
INSERT INTO contacts VALUES('c2','Marcus Johnson','+12125559876');
INSERT INTO contacts VALUES('c3','Emily Park','+13105553456');
INSERT INTO contacts VALUES('c4','David Kim','+14085557890');
INSERT INTO contacts VALUES('c5','Lisa Rodriguez','+12025551111');
INSERT INTO contacts VALUES('c6','Alex Thompson','+17185552222');
INSERT INTO contacts VALUES('c7','Rachel Green','+16505553333');
CREATE INDEX idx_messages_conv_ts ON messages(conversation_id, timestamp_ms);
CREATE INDEX idx_messages_ts ON messages(timestamp_ms DESC);
COMMIT;
