import os
import subprocess
import shutil

# Make frontend directory
frontend_dir = os.path.join(os.getcwd(), 'frontend')
if not os.path.exists(frontend_dir):
    os.makedirs(frontend_dir)

# Run npx command to initialize Vite React TypeScript project
print(f"Creating vite project in {frontend_dir}")
result = subprocess.run(
    ["npm.cmd", "create", "vite@latest", "./", "--", "--template", "react-ts"],
    cwd=frontend_dir,
    capture_output=True,
    text=True
)
print("STDOUT:", result.stdout)
print("STDERR:", result.stderr)

# Output package.json
pkg_json = os.path.join(frontend_dir, 'package.json')
if os.path.exists(pkg_json):
    print("package.json created successfully!")
    with open(pkg_json, 'r') as f:
        print(f.read())
else:
    print("Failed to create package.json")
