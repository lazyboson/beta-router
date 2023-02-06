### Beta-router
 ### Configuration 
  Appropriate CTI URL in .env File at /cmd/routerservice/routerservice.env
 ### Build 
    docker build -t router -f docker/Dockerfile .
 ### Run
    docker run  -d --env-file ./cmd/routerservice/routerservice.env --name router --rm -p 9495:9495 --network=host router 
