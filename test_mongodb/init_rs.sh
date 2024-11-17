#!/bin/bash

mongosh --host mongo:27017 <<EOF
  var cfg = {
    "_id": "rs0",
    "version": 1,
    "members": [
      {
        "_id": 0,
        "host": "127.0.0.1:27017",
      }
    ]
  };
  rs.initiate(cfg);
EOF