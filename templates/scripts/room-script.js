let socket;

function playCard(component) {
    const [kind, cardName] = component.alt.split('-');
    console.log({ kind, cardName });

    socket.send(JSON.stringify({
        command: 'play', kind, cardName,
    }));
}

function switchCard(component) {
    const [kind, cardName] = component.alt.split('-');
    console.log('test');

    socket.send(JSON.stringify({
        command: 'switch', kind, cardName,
    }));
}

function gameOff() {
    socket.send(JSON.stringify({
        command: 'gameOff',
    }))
}

function endTurn() {
    socket.send(JSON.stringify({
        command: 'endTurn',
    }))
}

window.onload = function () {
    const roomId = document.getElementById('roomId').value;
    console.log(`${roomId} trying to connect.`);
    socket = new WebSocket('wss://salty-fortress-42507.herokuapp.com/ws/'+roomId);

    socket.onopen = function(e) {
        console.log("[open] Connection established");
    };

    socket.onmessage = function(event) {
        console.log(`[message] Data received from server: ${event.data}`);

        const data = JSON.parse(event.data);

        if (data.command === 'start') {
            document.getElementById('cardsOnGround').innerHTML = '';
            document.getElementById('bottom-side').innerHTML = '';
            document.getElementById('turnScore').innerText = '0';
            const ground = document.createElement('img');
            ground.id = 'groundImg';
            ground.className = 'card cardWidth';
            ground.src = `/cards/${data.cardOnGround.Kind}-${data.cardOnGround.CardName}.svg`;
            ground.setAttribute('onclick', 'switchCard(this);');
            document.getElementById('cardsOnGround').appendChild(ground);

            const back = document.createElement('img');
            back.id = "groundBackImg";
            back.className = 'card cardWidth';
            back.src = `/cards/BACK.svg`;
            back.style = "position: relative; left: 0; transform: rotate(45deg); top: 50px;"
            document.getElementById('cardsOnGround').appendChild(back);

            data.cards.forEach((card) => {
                const cardImg = document.createElement('img');
                cardImg.className = 'card cardWidth deck';
                cardImg.src = `/cards/${card.Kind}-${card.CardName}.svg`;
                cardImg.setAttribute("alt",`${card.Kind}-${card.CardName}`);
                cardImg.setAttribute("onclick","playCard(this);");

                document.getElementById('bottom-side').appendChild(cardImg);
            });
        }

        if (data.command === 'play1') {
            if (data.isItMine === true) {
                Array.from(document.getElementsByClassName('deck')).forEach((img) => {
                    if (img.alt === `${data.moveMessage.kind}-${data.moveMessage.cardName}`) {
                        document.getElementById('ground').appendChild(img);
                    }
                });
            } else {
                const cardImg = document.createElement('img');
                cardImg.className = 'card cardWidth';
                cardImg.src = `/cards/${data.moveMessage.kind}-${data.moveMessage.cardName}.svg`;
                cardImg.setAttribute("alt",`${data.moveMessage.kind}-${data.moveMessage.cardName}`);
                document.getElementById('ground').appendChild(cardImg);
            }
        }

        if (data.command === 'play2') {
            if (data.isItMine === true) {
                Array.from(document.getElementsByClassName('deck')).forEach((img) => {
                    if (img.alt === `${data.moveMessage.kind}-${data.moveMessage.cardName}`) {
                        document.getElementById('ground').appendChild(img);
                    }
                });
            } else {
                const cardImg = document.createElement('img');
                cardImg.className = 'card cardWidth';
                cardImg.src = `/cards/${data.moveMessage.kind}-${data.moveMessage.cardName}.svg`;
                cardImg.setAttribute("alt",`${data.moveMessage.kind}-${data.moveMessage.cardName}`);
                document.getElementById('ground').appendChild(cardImg);
            }

            if (data.newCardName && data.newCardKind) {
                const cardImg = document.createElement('img');
                cardImg.className = 'card cardWidth deck';
                cardImg.src = `/cards/${data.newCardKind}-${data.newCardName}.svg`;
                cardImg.setAttribute("alt",`${data.newCardKind}-${data.newCardName}`);
                cardImg.setAttribute("onclick","playCard(this);");

                document.getElementById('bottom-side').appendChild(cardImg);
            }

            document.getElementById('turnScore').innerText = data.turnScore;

            setTimeout(() => {
                document.getElementById('ground').innerHTML = '';
            }, 1000);
        }

        if (data.command === "switch") {
            if (data.isItMine) {
                Array.from(document.getElementsByClassName('deck')).forEach((img) => {
                    if (img.alt === `${data.newCardKind}-9`) {
                        img.setAttribute('alt', `${data.newCardKind}-${data.newCardName}`);
                        img.src = `/cards/${data.newCardKind}-${data.newCardName}.svg`;

                        document.getElementById('groundImg').src = `/cards/${data.newCardKind}-9.svg`;
                        document.getElementById('groundImg').setAttribute("alt",`${data.newCardKind}-9`);
                    }
                });
            } else {
                document.getElementById('groundImg').src = `/cards/${data.newCardKind}-${data.newCardName}.svg`;
                document.getElementById('groundImg').setAttribute("alt",`${data.newCardKind}-${data.newCardName}`);
            }
        }

        if (data.command === 'removeGround') {
            document.getElementById('cardsOnGround').innerHTML = '';
        }

        if (data.command === 'endTurn') {
            document.getElementById('totalScore').innerText = data.totalScore;
            document.getElementById('enemyTotalScore').innerText = data.enemyTotalScore;
        }
    };

    socket.onclose = function(event) {
        if (event.wasClean) {
            console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
        } else {
            // e.g. server process killed or network down
            // event.code is usually 1006 in this case
            console.log('[close] Connection died');
        }
    };

    socket.onerror = function(error) {
        console.log(`[error] ${error.message}`);
    };
};