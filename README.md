# z2 - Zone 2 pace monitor

z2 is a [zone 2](https://peterattiamd.com/category/exercise/aerobic-zone-2-training/) pace monitor.

z2 typically runs on a raspberry pi and connects to fitness equipment via bluetooth. Dashboards can be viewed in your web
browser and are updated with low latency.

Currently supported equipment:

- Schwinn IC4/Bowflex C6 stationary bike
- Concept 2 rower
- Heart rate monitors

All metrics are recorded and stored for later review. Dashboards typically display the raw values as well as moving
averages of the metrics. This allows me to maintain a much steadier pace than by using the fitness equipment's built-in
displays.

In the future, z2 will calculate an allowable deviation from the target pace and use audio cues ("too fast" / "too
slow") to nudge you back on track. The current trend of fitness apps has you staring at screens and living in virtual
worlds. I want to take the exact opposite approach and remove the screen completely. I'd like to be able to work out
with no screen at all or even with my eyes closed. We spend so much of our lives looking at screens and this is an
opportunity to take back a small amount of that time (while still using technology to monitor your pace and track your
stats).

## Bike screenshot

![app screenshot](https://minor-industries.sfo2.digitaloceanspaces.com/sw/z2_screenshot_02.png)

## Rower screenshot

Coming soon.

## Roadmap

- Calculate and display target zone
- Audio cues
- Randomized workout duration