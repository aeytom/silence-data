# Silence Connected API client

- reads data from https://api.connectivity.silence.eco/api/v1/
- used API methods; see (silence.go)[silence.go]
    - login, refreshToken – authorisation
    - me – user profile, avatar (not fully implemented)
    - me/scooters – scooter, battery information
    - scooters/*/trips – trip information
- write infos to influxdb2 bucket
- simple grafana dashboard (not a template yet)