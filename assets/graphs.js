Dygraph.onDOMready(function onDOMready() {
    fetch('data.json' + window.location.search)
        .then(response => response.json())
        .then(response => {
            let data = response.rows.map(value => [
                new Date(value[0]),
                value[1]
            ]);
            window.g = new Dygraph(
                // containing div
                document.getElementById("graphdiv"),
                // CSV or path to a CSV file.
                data,
                {
                    // dateWindow: [t0, t1],
                    title: "Title",
                    ylabel: "ylabel",
                    labels: ["X", "Y"]
                }
            );
        })
        .catch(error => console.error('Error:', error));


    const url = `ws://${window.location.hostname}:${window.location.port}/ws`;
    const ws = new WebSocket(url);
    ws.onopen = console.log;
    ws.onmessage = message => {
        const msg = JSON.parse(message.data);
        console.log(msg);
    };

    // let data = "data.csv" + window.location.search;
    //
    // const t1 = new Date();
    // const t0 = new Date(t1);
    // t0.setMonth(t0.getMonth() - 3);
    // t1.setDate(t0.getDate() + 1); // maybe better to use some padding options here instead
});