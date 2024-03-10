const mapDate = value => [new Date(value[0]), value[1]];

function makeGraph(elem) {
    let g;
    let data;

    fetch('data.json' + window.location.search)
        .then(response => response.json())
        .then(response => {
            data = response.rows.map(mapDate);
            g = new Dygraph(// containing div
                elem,
                data, {
                    // dateWindow: [t0, t1],
                    title: "Title",
                    ylabel: "ylabel",
                    labels: ["X", "Y"],
                });
        })
        .catch(error => console.error('Error:', error));

    const url = `ws://${window.location.hostname}:${window.location.port}/ws`;
    const ws = new WebSocket(url);
    ws.onopen = console.log;
    ws.onmessage = message => {
        const msg = JSON.parse(message.data);
        const rows = msg.rows.map(mapDate);

        const last = rows[rows.length - 1];
        const t1 = new Date(last[0]);
        const t0 = new Date(t1);
        t0.setMinutes(t0.getMinutes() - 5);

        data.push(...rows);
        g.updateOptions({
            file: window.data,
            dateWindow: [t0, t1],
        });
    };

    // let data = "data.csv" + window.location.search;
    //
    // const t1 = new Date();
    // const t0 = new Date(t1);
    // t0.setMonth(t0.getMonth() - 3);
    // t1.setDate(t0.getDate() + 1); // maybe better to use some padding options here instead
}

Dygraph.onDOMready(function onDOMready() {
    makeGraph(document.getElementById("graphdiv0"));
    makeGraph(document.getElementById("graphdiv1"));
});
