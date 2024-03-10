const mapDate = value => [new Date(value[0]), value[1]];

function makeGraph(elem, opts) {
    let g;
    let data;

    fetch('data.json' + window.location.search)
        .then(response => response.json())
        .then(response => {
            const t1 = new Date(response.now);
            const t0 = new Date(t1);
            t0.setMinutes(t0.getMinutes() - 5);

            data = response.rows.map(mapDate);
            g = new Dygraph(// containing div
                elem,
                data, {
                    // dateWindow: [t0, t1],
                    title: opts.title,
                    ylabel: opts.ylabel,
                    labels: ["X", "Y"],
                    dateWindow: [t0, t1],
                });
        })
        .then(() => {
            const url = `ws://${window.location.hostname}:${window.location.port}/ws`;
            const ws = new WebSocket(url);
            ws.onmessage = message => {
                console.log("message");
                const msg = JSON.parse(message.data);
                const rows = msg.rows.map(mapDate);

                const last = rows[rows.length - 1];
                const t1 = new Date(last[0]);
                const t0 = new Date(t1);
                t0.setMinutes(t0.getMinutes() - 5);

                data.push(...rows);
                console.log(data.length);
                g.updateOptions({
                    file: data,
                    dateWindow: [t0, t1],
                });
            };
            ws.onopen = event => {
                setTimeout(function () {
                    ws.send(JSON.stringify({series: opts.series}));
                })
            }
        })
        .catch(error => console.error('Error:', error));


    // let data = "data.csv" + window.location.search;
    //
    // const t1 = new Date();
    // const t0 = new Date(t1);
    // t0.setMonth(t0.getMonth() - 3);
    // t1.setDate(t0.getDate() + 1); // maybe better to use some padding options here instead
}

Dygraph.onDOMready(function onDOMready() {
    makeGraph(document.getElementById("graphdiv0"), {
        series: "bike_instant_speed",
        title: "Speed",
        ylabel: "speed (km/h)"
    });
    makeGraph(document.getElementById("graphdiv1"), {
        series: "bike_instant_power",
        title: "Power",
        ylabel: "power (watts)"
    });
});
