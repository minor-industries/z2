const mapDate = value => [new Date(value[0]), value[1]];

Dygraph.onDOMready(function onDOMready() {
    fetch('data.json' + window.location.search)
        .then(response => response.json())
        .then(response => {
            window.data = response.rows.map(mapDate);
            window.g = new Dygraph(// containing div
                document.getElementById("graphdiv"), // CSV or path to a CSV file.
                window.data, {
                    // dateWindow: [t0, t1],
                    title: "Title",
                    ylabel: "ylabel",
                    labels: ["X", "Y"],
                    xRangePad: 100,
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

        window.data.push(...rows);
        window.g.updateOptions({
            file: window.data,
            dateWindow: [t0, t1],
            xRangePad: 100,
        });
    };

    // let data = "data.csv" + window.location.search;
    //
    // const t1 = new Date();
    // const t0 = new Date(t1);
    // t0.setMonth(t0.getMonth() - 3);
    // t1.setDate(t0.getDate() + 1); // maybe better to use some padding options here instead
});
