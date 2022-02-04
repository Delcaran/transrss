docker build -t alpine/transrss:latest .
docker run -v %cd%:/ext alpine/transrss:latest cp /app/transrss /ext