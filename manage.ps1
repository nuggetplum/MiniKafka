param (
    [string]$Command = "run"
)

Switch ($Command) {
    "build" {
        Write-Host "Building..."
        go build -o bin/minikafka.exe cmd/server/main.go
    }
    "run" {
        Write-Host "Running..."
        go run cmd/server/main.go
    }
    "test" {
        Write-Host "Testing..."
        go test -v ./...
    }
    "clean" {
        Write-Host "Cleaning..."
        go clean
        if (Test-Path bin) { Remove-Item -Recurse -Force bin }
        if (Test-Path prog_log) { Remove-Item -Recurse -Force prog_log }
    }
    Default {
        Write-Host "Unknown command. Use: build, run, test, clean"
    }
}