# b-gals
B-gals sit prosjekt i TTK4145

### Hvordan bruke git:
`git add .`\ 
`git commit -m "Jeg fiksa det"`\
`git push`

### Når du kommer på sanntidslaben:
`git pull`
//eller bare når du skal oppdatere og se hva de andre har gjort



Sjekke om vi er konsekevente i bruk av små og store forbokstaver (public og private)

### Vi er ganske smarte jenter
Ja det er vi


### Hvordan kjøre heisene
Run the build.sh script to build the program <br>
`bash build.sh`

På hver av pcene så kjører dere en av 3 config-filene, som er elevator1.con, elevator2.con og elevator3.con. Eksempel på pc1 er: <br>
`./heis config/elevator1.con`

## How to Run on physical elevator
```bash
cd ~
git clone https://github.com/uyeennh/b-gals.git
cd b-gals/Heis2
bash build.sh
./heis config/elevator1.con   # change number per PC (1, 2, or 3)
```

## How to Run with Simulator 

Start a simulator for each elevator in separate terminals:
```bash
cd Simulator-v2-master\(2\)/Simulator-v2-master
./SimElevatorServer --port 15657   # Terminal 1 — for elevator 1
./SimElevatorServer --port 15658   # Terminal 2 — for elevator 2
./SimElevatorServer --port 15659   # Terminal 3 — for elevator 3
```

Then start the elevators in separate terminals:
```bash
./heis config/elevator1.con   # Terminal 4
./heis config/elevator2.con   # Terminal 5
./heis config/elevator3.con   # Terminal 6
```

### Simulator keyboard controls
```
q w e r    — Hall Up   buttons for floors 0, 1, 2, 3
a s d f    — Hall Down buttons for floors 0, 1, 2, 3
z x c v    — Cab       buttons for floors 0, 1, 2, 3
t          — Stop button
g          — Obstruction on/off
