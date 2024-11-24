import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

def process_csv(filename):
    # Read and filter data
    contents = open(filename, "r").read().split('\n')
    filtered_lines = [line for line in contents if line.strip() and 'BenchmarkGol' in line]

    # Create DataFrame
    data = []
    for line in filtered_lines:
        parts = line.strip().split(',')
        if len(parts) >= 3:
            thread_count = int(parts[0].split('-')[1])
            time_ns = float(parts[2])
            data.append([thread_count, time_ns])

    df = pd.DataFrame(data, columns=['threads', 'time_ns'])
    df['time'] = df['time_ns'] / 1e9  # Convert to seconds
    return df

# Read and process data
benchmark_data = process_csv("parallel_results.csv")

# Calculate y-axis limits with 5% padding
y_min = benchmark_data['time'].min() * 0.95  # 5% lower
y_max = benchmark_data['time'].max() * 1.05  # 5% higher

# Create line plot
plt.figure(figsize=(8, 5))  # Reduced width from 10 to 8
ax = sns.lineplot(data=benchmark_data, x='threads', y='time', marker='o')  # Added markers to emphasize data points

plt.title('Parallel Implementation on Distributed System Performance')
plt.xlabel('Number of Threads in Each Worker')
plt.ylabel('Time (seconds)')
plt.ylim(y_min, y_max)
plt.grid(True)
plt.tight_layout()
plt.savefig('parallel_benchmark.png', dpi=300)
plt.close()