<?php

/**
 * 检查前端国际化翻译中未使用的键
 * 扫描frontend/src目录下的所有.tsx和.ts文件，查找未使用的翻译键
 */

class I18nChecker {
    private $frontendPath;
    private $i18nPath;
    private $srcPath;
    private $usedKeys = [];
    private $allKeys = [];
    
    public function __construct($basePath = '.') {
        $this->frontendPath = rtrim($basePath, '/') . '/frontend';
        $this->i18nPath = $this->frontendPath . '/src/i18n/locales/en-US';
        $this->srcPath = $this->frontendPath . '/src';
    }
    
    /**
     * 主执行方法
     */
    public function run() {
        echo "🔍 检查前端国际化翻译中未使用的键...\n\n";
        
        // 检查路径是否存在
        if (!is_dir($this->i18nPath)) {
            die("❌ 错误: 翻译文件目录不存在: {$this->i18nPath}\n");
        }
        
        if (!is_dir($this->srcPath)) {
            die("❌ 错误: 前端源码目录不存在: {$this->srcPath}\n");
        }
        
        // 1. 读取所有翻译键
        $this->loadAllTranslationKeys();
        
        // 2. 扫描源码文件，查找使用的翻译键
        $this->scanSourceFiles();
        
        // 3. 找出未使用的键
        $this->findUnusedKeys();
    }
    
    /**
     * 加载所有翻译键
     */
    private function loadAllTranslationKeys() {
        echo "📖 读取翻译文件...\n";
        
        // 文件名到命名空间的映射
        $namespaceMapping = [
            'adminLogs' => 'adminLogs',
            'api' => 'api',
            'auth' => 'auth',
            'common' => 'common',
            'dashboard' => 'dashboard',
            'devEnvironments' => 'devEnvironments',
            'errors' => 'errors',
            'gitCredentials' => 'gitCredentials',
            'gitDiff' => 'gitDiff',
            'navigation' => 'navigation',
            'projects' => 'projects',
            'systemConfig' => 'systemConfig',
            'taskConversations' => 'taskConversations',
            'tasks' => 'tasks'
        ];
        
        $files = glob($this->i18nPath . '/*.json');
        
        foreach ($files as $file) {
            $fileName = basename($file, '.json');
            $namespace = $namespaceMapping[$fileName] ?? $fileName;
            echo "   - {$fileName}.json (命名空间: {$namespace})\n";
            
            $content = file_get_contents($file);
            $data = json_decode($content, true);
            
            if ($data === null) {
                echo "   ⚠️  警告: 无法解析JSON文件 {$file}\n";
                continue;
            }
            
            $this->extractKeys($data, $namespace);
        }
        
        echo "✅ 总共找到 " . count($this->allKeys) . " 个翻译键\n\n";
    }
    
    /**
     * 递归提取所有翻译键
     */
    private function extractKeys($data, $prefix = '') {
        foreach ($data as $key => $value) {
            $fullKey = $prefix ? $prefix . '.' . $key : $key;
            
            if (is_array($value)) {
                $this->extractKeys($value, $fullKey);
            } else {
                // 使用键来避免重复
                $this->allKeys[$fullKey] = true;
            }
        }
    }
    
    /**
     * 扫描源码文件
     */
    private function scanSourceFiles() {
        echo "🔍 扫描源码文件中的翻译键使用...\n";
        
        $iterator = new RecursiveIteratorIterator(
            new RecursiveDirectoryIterator($this->srcPath)
        );
        
        $fileCount = 0;
        foreach ($iterator as $file) {
            if ($file->isFile() && preg_match('/\.(tsx?|jsx?)$/', $file->getFilename())) {
                $this->scanFile($file->getPathname());
                $fileCount++;
            }
        }
        
        echo "✅ 扫描了 {$fileCount} 个文件，找到 " . count($this->usedKeys) . " 个使用的翻译键\n\n";
    }
    
    /**
     * 扫描单个文件
     */
    private function scanFile($filePath) {
        $content = file_get_contents($filePath);
        
        // 移除单行和多行注释，避免误匹配注释中的内容
        $content = preg_replace('/\/\*[\s\S]*?\*\//', '', $content);
        $content = preg_replace('/\/\/.*$/', '', $content);
        
        // 匹配 t("key") 或 t('key') 的模式，支持多行和空白字符
        // 使用 DOTALL 修饰符让 . 匹配换行符
        preg_match_all('/\bt\(\s*["\']([^"\']+)["\']\s*(?:,[\s\S]*?)?\)/s', $content, $matches);
        
        if (!empty($matches[1])) {
            foreach ($matches[1] as $key) {
                // 使用键来避免重复
                $this->usedKeys[$key] = true;
            }
        }
    }
    
    /**
     * 查找未使用的键
     */
    private function findUnusedKeys() {
        echo "🔎 分析未使用的翻译键...\n\n";
        
        // 获取所有键和已使用键的数组
        $allKeysArray = array_keys($this->allKeys);
        $usedKeysArray = array_keys($this->usedKeys);
        
        $unusedKeys = array_diff($allKeysArray, $usedKeysArray);
        
        if (empty($unusedKeys)) {
            echo "🎉 太棒了！所有翻译键都被使用了！\n";
            return;
        }
        
        // 命名空间到文件名的反向映射
        $fileNameMapping = [
            'adminLogs' => 'adminLogs',
            'api' => 'api',
            'auth' => 'auth',
            'common' => 'common',
            'dashboard' => 'dashboard',
            'devEnvironments' => 'devEnvironments',
            'errors' => 'errors',
            'gitCredentials' => 'gitCredentials',
            'gitDiff' => 'gitDiff',
            'navigation' => 'navigation',
            'projects' => 'projects',
            'systemConfig' => 'systemConfig',
            'taskConversations' => 'taskConversations',
            'tasks' => 'tasks'
        ];
        
        // 按模块分组显示未使用的键
        $groupedUnused = [];
        foreach ($unusedKeys as $key) {
            $parts = explode('.', $key);
            $namespace = $parts[0];
            $fileName = $fileNameMapping[$namespace] ?? $namespace;
            $remainingKey = implode('.', array_slice($parts, 1));
            
            if (!isset($groupedUnused[$fileName])) {
                $groupedUnused[$fileName] = [];
            }
            $groupedUnused[$fileName][] = $remainingKey;
        }
        
        echo "❌ 发现 " . count($unusedKeys) . " 个未使用的翻译键：\n\n";
        
        foreach ($groupedUnused as $module => $keys) {
            echo "📁 {$module}.json (" . count($keys) . " 个未使用):\n";
            foreach ($keys as $key) {
                echo "   - {$key}\n";
            }
            echo "\n";
        }
        
        // 统计信息
        echo "📊 统计信息:\n";
        echo "   - 总翻译键数: " . count($this->allKeys) . "\n";
        echo "   - 已使用键数: " . count($this->usedKeys) . "\n";
        echo "   - 未使用键数: " . count($unusedKeys) . "\n";
        echo "   - 使用率: " . round((count($this->usedKeys) / count($this->allKeys)) * 100, 2) . "%\n\n";
        
        // 验证数量
        $total = count($this->allKeys);
        $used = count($this->usedKeys);
        $unused = count($unusedKeys);
        
        echo "🔍 数量验证:\n";
        echo "   - 总键数: {$total}\n";
        echo "   - 已使用: {$used}\n";
        echo "   - 未使用: {$unused}\n";
        echo "   - 验证: {$used} + {$unused} = " . ($used + $unused) . " (应该等于 {$total})\n";
        
        // 调试：检查是否有重复的已使用键
        $allKeysArray = array_keys($this->allKeys);
        $usedKeysArray = array_keys($this->usedKeys);
        
        // 检查已使用键中是否有不在总键列表中的
        $invalidUsedKeys = array_diff($usedKeysArray, $allKeysArray);
        if (!empty($invalidUsedKeys)) {
            echo "⚠️  发现 " . count($invalidUsedKeys) . " 个无效的已使用键（不在翻译文件中）:\n";
            foreach (array_slice($invalidUsedKeys, 0, 10) as $key) {
                echo "   - {$key}\n";
            }
            if (count($invalidUsedKeys) > 10) {
                echo "   - ... 还有 " . (count($invalidUsedKeys) - 10) . " 个\n";
            }
        }
        
        // 重新计算正确的统计
        $validUsedKeys = array_intersect($usedKeysArray, $allKeysArray);
        $actualUnused = array_diff($allKeysArray, $validUsedKeys);
        
        echo "\n📊 修正后的统计:\n";
        echo "   - 总翻译键数: " . count($allKeysArray) . "\n";
        echo "   - 有效使用键数: " . count($validUsedKeys) . "\n";
        echo "   - 实际未使用键数: " . count($actualUnused) . "\n";
        echo "   - 使用率: " . round((count($validUsedKeys) / count($allKeysArray)) * 100, 2) . "%\n\n";
        
        // 生成清理建议
        echo "💡 清理建议:\n";
        echo "   可以考虑删除这些未使用的翻译键以减少文件大小\n";
        echo "   删除前请确认这些键确实不会在动态生成的场景中使用\n";
    }
}

// 执行检查
try {
    $checker = new I18nChecker();
    $checker->run();
} catch (Exception $e) {
    echo "❌ 执行出错: " . $e->getMessage() . "\n";
    exit(1);
}

?>