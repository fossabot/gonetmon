# Todo

## Requirements

- [x] Sniff/Capture Traffic
  - [x] On all network interfaces
  - [x] Add filters :
    - [x] TCP and port 80
- [x] Time based console display with traffic information
  - [ ] Most hits on website sections
- [x] Alert when traffic over past n minutes hits a threshold
  - [x] Inform when recovered
- [ ] Tests

### Remain

- Hit Analysis
- Report building
- Hit display

### Error handling

- Thoroughly check if all errors are handled

### Other Todo

- [ ] Tests :
  - Platforms
    - [ ] Linux
    - [ ] Macos
    - [ ] maybe Windows ?
  - With real world traffic
    - [ ] Simulate with user behaviour ?
    - [ ] Automated crawler ?
  - Ad-hoc tests
    - [ ] Tailor specific requests to trigger hit detection and alerts

- [x] Code quality
  - [x] goreportcard
  - [x] Codacy
  - [x] Sonar
- [x] Vulnerability checks
  - [x] Snyk
