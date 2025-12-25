import csv

import matplotlib.pyplot as plt

x = []
y = []

with open("waveprocess_amp.csv") as f:
    reader = csv.DictReader(f)
    for row in reader:
        x.append(int(row["x"]))
        y.append(float(row["f"]))

plt.plot(x, y, marker="o")
plt.xlabel("sample")
plt.ylabel("amplitude")
plt.title("Amplitude Graph")
plt.grid(True)
plt.savefig("waveprocess_amp.png")
# plt.show()
