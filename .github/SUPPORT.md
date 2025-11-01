# Getting Help with Juleson

Thank you for using Juleson! Here are the best ways to get help.

## 📖 Documentation

Before asking for help, please check our comprehensive documentation:

- **[README](../README.md)** - Quick start and overview
- **[MCP Server Usage](../docs/MCP_SERVER_USAGE.md)** - Model Context Protocol integration
- **[Template System](../docs/Y2Q2_TEMPLATE_SYSTEM.md)** - Template documentation
- **[GitHub Actions Guide](../docs/GITHUB_ACTIONS_GUIDE.md)** - CI/CD integration
- **[Contributing Guide](../CONTRIBUTING.md)** - Development setup

## 💬 Community Support

### GitHub Discussions

For general questions, ideas, and community discussion:

👉 **[GitHub Discussions](https://github.com/SamyRai/Juleson/discussions)**

**Use Discussions for:**

- ❓ Questions about usage
- 💡 Feature ideas and suggestions
- 🎉 Show & Tell - Share what you've built
- 💬 General conversation

### GitHub Issues

For bug reports and specific technical problems:

👉 **[GitHub Issues](https://github.com/SamyRai/Juleson/issues)**

**Use Issues for:**

- 🐛 [Report a Bug](https://github.com/SamyRai/Juleson/issues/new?template=bug_report.md)
- ✨ [Request a Feature](https://github.com/SamyRai/Juleson/issues/new?template=feature_request.md)

Before opening an issue:

1. Search [existing issues](https://github.com/SamyRai/Juleson/issues) to avoid duplicates
2. Check [closed issues](https://github.com/SamyRai/Juleson/issues?q=is%3Aissue+is%3Aclosed) for solutions
3. Prepare reproduction steps and environment details

## 🔍 Troubleshooting

### Common Issues

**Jules CLI not working:**

```bash
# Check your configuration
cat configs/Juleson.yaml

# Verify API key
echo $JULES_API_KEY

# Test connection
./bin/juleson --help
```

**MCP Server connection issues:**

- Ensure stdio transport is configured
- Check Claude Desktop config: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Verify the MCP server path is correct

**Template execution fails:**

- Verify internet connection (required for Jules API)
- Check API key permissions
- Review template syntax with `juleson template view <template-name>`

### Logs and Debugging

Enable verbose logging:

```bash
export JULES_LOG_LEVEL=debug
./bin/juleson <command>
```

## 📧 Direct Support

### Email Support

For private inquiries or security issues:

- **Security vulnerabilities**: <security@glpx.pro> (see [SECURITY.md](../SECURITY.md))
- **General inquiries**: Create a [Discussion](https://github.com/SamyRai/Juleson/discussions)

**Response times:**

- Security issues: 48 hours
- General questions: Best effort (usually within a week)
- Community discussions: Variable

## 🤝 Contributing

Want to contribute? That's great! Please read:

- [Contributing Guidelines](../CONTRIBUTING.md)
- [Code of Conduct](../CODE_OF_CONDUCT.md)

## 📚 Additional Resources

- **Jules AI Documentation**: [jules.ai/docs](https://jules.ai/docs)
- **MCP Protocol**: [Model Context Protocol](https://modelcontextprotocol.io/)
- **Go Documentation**: [golang.org](https://go.dev/)

## ⚡ Quick Links

| I want to... | Go here... |
|--------------|------------|
| Ask a question | [Discussions](https://github.com/SamyRai/Juleson/discussions) |
| Report a bug | [New Issue](https://github.com/SamyRai/Juleson/issues/new?template=bug_report.md) |
| Request a feature | [New Issue](https://github.com/SamyRai/Juleson/issues/new?template=feature_request.md) |
| Read documentation | [README](../README.md) or [/docs](../docs/) |
| Contribute code | [Contributing Guide](../CONTRIBUTING.md) |
| Report security issue | [SECURITY.md](../SECURITY.md) |

---

**Remember**: The community is here to help! Don't hesitate to ask questions. 🎉
